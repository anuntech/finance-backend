package transaction

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"slices"

	"github.com/anuntech/finance-backend/internal/domain/models"
	"github.com/anuntech/finance-backend/internal/domain/usecase"
	infraHelpers "github.com/anuntech/finance-backend/internal/infra/db/mongodb/helpers"
	"github.com/anuntech/finance-backend/internal/infra/db/mongodb/repositories/transaction_repository"
	"github.com/anuntech/finance-backend/internal/infra/db/mongodb/repositories/workspace_repository/member_repository"
	"github.com/anuntech/finance-backend/internal/presentation/helpers"
	presentationProtocols "github.com/anuntech/finance-backend/internal/presentation/protocols"
	"github.com/anuntech/finance-backend/internal/utils"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	"github.com/xuri/excelize/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type FindMemberByEmailRepository interface {
	FindByEmailAndWorkspaceId(email string, workspaceId primitive.ObjectID) (*models.Member, error)

	FindByNameAndWorkspaceId(name string, workspaceId primitive.ObjectID) (*models.Member, error)
}

type FindCustomFieldByNameRepository interface {
	FindByNameAndWorkspaceId(name string, workspaceId primitive.ObjectID) (*models.CustomField, error)
}

type CreateAccountRepository interface {
	Create(account *models.Account) (*models.Account, error)
}

type CreateCategoryRepository interface {
	Create(category *models.Category) (*models.Category, error)
}

type ImportTransactionController struct {
	Validate                      *validator.Validate
	Translator                    ut.Translator
	CreateTransactionRepository   usecase.CreateTransactionRepository
	FindMemberByIdRepository      *member_repository.FindMemberByIdRepository
	FindAccountByIdRepository     usecase.FindAccountByIdRepository
	FindCategoryByIdRepository    usecase.FindCategoryByIdRepository
	FindCustomFieldByIdRepository usecase.FindCustomFieldByIdRepository

	FindAccountByNameRepository         usecase.FindAccountByNameAndWorkspaceIdRepository
	FindCategoryByNameAndTypeRepository usecase.FindCategoryByNameAndTypeRepository
	FindMemberByEmailRepository         FindMemberByEmailRepository
	FindCustomFieldByNameRepository     FindCustomFieldByNameRepository

	CreateAccountRepository  CreateAccountRepository
	CreateCategoryRepository CreateCategoryRepository
	FindBankByNameRepository usecase.FindBankByNameRepository

	categoryCache    categoryCache
	accountCache     accountCache
	memberCache      memberCache
	customFieldCache customFieldCache
	bankCache        bankCache
}

func NewImportTransactionController(
	findMemberByIdRepository *member_repository.FindMemberByIdRepository,
	createTransactionRepository *transaction_repository.CreateTransactionRepository,
	findAccountByIdRepository usecase.FindAccountByIdRepository,
	findCategoryByIdRepository usecase.FindCategoryByIdRepository,
	findCustomFieldByIdRepository usecase.FindCustomFieldByIdRepository,
	findAccountByNameRepository usecase.FindAccountByNameAndWorkspaceIdRepository,
	findCategoryByNameAndTypeRepository usecase.FindCategoryByNameAndTypeRepository,
	findMemberByEmailRepository FindMemberByEmailRepository,
	findCustomFieldByNameRepository FindCustomFieldByNameRepository,
	createAccountRepository CreateAccountRepository,
	createCategoryRepository CreateCategoryRepository,
	findBankByNameRepository usecase.FindBankByNameRepository,
) *ImportTransactionController {
	validate := validator.New(validator.WithRequiredStructEnabled())

	return &ImportTransactionController{
		Validate:                            validate,
		FindMemberByIdRepository:            findMemberByIdRepository,
		CreateTransactionRepository:         createTransactionRepository,
		FindAccountByIdRepository:           findAccountByIdRepository,
		FindCategoryByIdRepository:          findCategoryByIdRepository,
		FindCustomFieldByIdRepository:       findCustomFieldByIdRepository,
		FindAccountByNameRepository:         findAccountByNameRepository,
		FindCategoryByNameAndTypeRepository: findCategoryByNameAndTypeRepository,
		FindMemberByEmailRepository:         findMemberByEmailRepository,
		FindCustomFieldByNameRepository:     findCustomFieldByNameRepository,
		CreateAccountRepository:             createAccountRepository,
		CreateCategoryRepository:            createCategoryRepository,
		FindBankByNameRepository:            findBankByNameRepository,
		categoryCache:                       categoryCache{items: make(map[cacheKey]*models.Category)},
		accountCache:                        accountCache{items: make(map[cacheKey]*models.Account)},
		memberCache:                         memberCache{items: make(map[cacheKey]*models.Member)},
		customFieldCache:                    customFieldCache{items: make(map[cacheKey]*models.CustomField)},
		bankCache:                           bankCache{items: make(map[string]*models.Bank)},
	}
}

type TransactionImportItem struct {
	Name        string `json:"name" validate:"required,min=2,max=30"`
	Description string `json:"description" validate:"omitempty,max=255"`
	Invoice     string `json:"invoice" validate:"omitempty,min=2,max=50"`
	Type        string `json:"type" validate:"required,oneof=EXPENSE RECIPE"`
	Supplier    string `json:"supplier" validate:"omitempty,min=3,max=30"`
	AssignedTo  string `json:"assignedTo" validate:"required,email"`
	Balance     struct {
		Value              float64 `json:"value" validate:"required,min=0.01"`
		Discount           float64 `json:"discount" validate:"omitempty,min=0.01"`
		Interest           float64 `json:"interest" validate:"omitempty,min=0.01"`
		DiscountPercentage float64 `json:"discountPercentage" validate:"omitempty,min=0.01,max=100"`
		InterestPercentage float64 `json:"interestPercentage" validate:"omitempty,min=0.01,max=100"`
	} `json:"balance" validate:"required"`
	Frequency      string `json:"frequency" validate:"oneof=DO_NOT_REPEAT RECURRING REPEAT"`
	RepeatSettings struct {
		InitialInstallment time.Month `json:"initialInstallment" validate:"min=1"`
		Count              int        `json:"count" validate:"min=2,max=367"`
		Interval           string     `json:"interval" validate:"oneof=DAILY WEEKLY MONTHLY QUARTERLY YEARLY CUSTOM"`
		CustomDay          int        `json:"customDay" validate:"omitempty,required_if=Interval CUSTOM"`
	} `json:"repeatSettings" validate:"excluded_if=Frequency DO_NOT_REPEAT,excluded_if=Frequency RECURRING,required_if=Frequency REPEAT,omitempty"`
	DueDate     string  `json:"dueDate" validate:"required,datetime=2006-01-02T15:04:05Z"`
	IsConfirmed bool    `json:"isConfirmed"`
	Category    *string `json:"categoryId" validate:"omitempty"`
	SubCategory *string `json:"subCategoryId" validate:"omitempty"`
	Tags        []struct {
		Tag    string `json:"tag" validate:"omitempty"`
		SubTag string `json:"subTag" validate:"omitempty"`
	} `json:"tags" validate:"omitempty"`
	CustomFields []struct {
		CustomField string `json:"customField" validate:"required"`
		Value       string `json:"value" validate:"required,max=100"`
	} `json:"customFields" validate:"omitempty"`
	Account          string  `json:"accountId"`
	RegistrationDate string  `json:"registrationDate" validate:"required,datetime=2006-01-02T15:04:05Z"`
	ConfirmationDate *string `json:"confirmationDate" validate:"excluded_if=IsConfirmed false,required_if=IsConfirmed true,omitempty,datetime=2006-01-02T15:04:05Z"`
}

type ImportTransactionBody struct {
	Transactions []TransactionImportItem `json:"transactions" validate:"required,dive"`
}

func (c *ImportTransactionController) Handle(r presentationProtocols.HttpRequest) *presentationProtocols.HttpResponse {
	transactions, err := c.ParseMultipartAndMap(r.Req)
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "error parsing multipart and mapping: " + err.Error(),
		}, http.StatusBadRequest)
	}

	validationErrors := []map[string]any{}

	transactions, err = c.ParseAllDatesAndTypes(transactions)
	if err != nil {
		if !strings.Contains(err.Error(), "erros na análise de datas:") {
			return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
				Error: "error parsing all dates: " + err.Error(),
			}, http.StatusBadRequest)
		}

		errorMsg := err.Error()

		if idx := strings.Index(errorMsg, "erros na análise de datas: "); idx >= 0 {
			errorMsg = errorMsg[idx+len("erros na análise de datas: "):]
		}

		dateErrors := strings.Split(errorMsg, "; ")

		for _, dateError := range dateErrors {
			re := regexp.MustCompile(`linha (\d+)`)
			matches := re.FindStringSubmatch(dateError)
			if len(matches) >= 2 {
				idx, err := strconv.Atoi(matches[1])
				if err == nil {
					validationErrors = append(validationErrors, map[string]any{
						"line":  idx,
						"error": dateError,
					})
				} else {
					validationErrors = append(validationErrors, map[string]any{
						"line":  0,
						"error": dateError,
					})
				}
			} else {
				validationErrors = append(validationErrors, map[string]any{
					"line":  0,
					"error": dateError,
				})
			}
		}
	}

	body := &ImportTransactionBody{
		Transactions: transactions,
	}

	if err := c.Validate.Struct(body); err != nil {

		if errs, ok := err.(validator.ValidationErrors); ok {
			for _, e := range errs {

				field := e.Namespace()
				index := 0

				re := regexp.MustCompile(`Transactions\[(\d+)\]`)
				matches := re.FindStringSubmatch(field)
				if len(matches) >= 2 {
					if idx, err := strconv.Atoi(matches[1]); err == nil {
						index = idx
					}
				}

				// Skip ConfirmationDate validation errors
				if strings.Contains(field, "ConfirmationDate") {
					continue
				}

				errorMsg := c.translateValidationError(e)
				validationErrors = append(validationErrors, map[string]any{
					"line":  index + 2,
					"error": errorMsg,
				})
			}
		} else {

			validationErrors = append(validationErrors, map[string]any{
				"line":  0,
				"error": "Erro de validação: " + err.Error(),
			})
		}
	}

	userID := r.Header.Get("userId")
	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		validationErrors = append(validationErrors, map[string]any{
			"line":  0,
			"error": "ID de usuário inválido",
		})
	}

	workspaceIdStr := r.Header.Get("workspaceId")
	workspaceId, err := primitive.ObjectIDFromHex(workspaceIdStr)
	if err != nil {
		validationErrors = append(validationErrors, map[string]any{
			"line":  0,
			"error": "ID de workspace inválido",
		})
		return helpers.CreateResponse(map[string]any{
			"total":  len(validationErrors),
			"errors": validationErrors,
		}, http.StatusBadRequest)
	}

	LIMIT := 15000
	if len(body.Transactions) > LIMIT {
		validationErrors = append(validationErrors, map[string]any{
			"line":  0,
			"error": "Máximo de " + strconv.Itoa(LIMIT) + " transações por importação",
		})
	}

	memberEmails := map[string]bool{}
	for _, tx := range body.Transactions {
		if tx.AssignedTo != "" {
			memberEmails[tx.AssignedTo] = true
		}
	}

	missingMembers := []string{}
	for email := range memberEmails {
		member, err := c.memberCache.getByEmail(email, workspaceId, c.FindMemberByEmailRepository.FindByEmailAndWorkspaceId)
		if err != nil || member == nil {
			missingMembers = append(missingMembers, email)
		}
	}

	if len(missingMembers) > 0 && float64(len(missingMembers))/float64(len(memberEmails)) > 0.5 {
		for i, tx := range body.Transactions {
			if slices.Contains(missingMembers, tx.AssignedTo) {
				validationErrors = append(validationErrors, map[string]any{
					"line":  i + 2,
					"error": "Usuário não encontrado: " + tx.AssignedTo + ". Verifique se este email está cadastrado em seu workspace.",
				})
			}
		}

		totalErrors := len(validationErrors)
		if totalErrors > 50 {
			validationErrors = validationErrors[0:50]
		}

		return helpers.CreateResponse(map[string]any{
			"total":  totalErrors,
			"errors": validationErrors,
		}, http.StatusBadRequest)
	}

	var wg sync.WaitGroup
	importedTransactions := make([]*models.Transaction, len(body.Transactions))
	type errorInfo struct {
		index int
		err   error
	}
	errs := make(chan errorInfo, len(body.Transactions))
	createdTransactions := make([]*models.Transaction, len(body.Transactions))
	const workers = 50
	sem := make(chan struct{}, workers)

	errTypeCount := make(map[string]int)
	var errMutex sync.Mutex

	earlyTerminate := make(chan bool, 1)
	stopProcessing := false
	var stopMutex sync.Mutex

	for i, txImport := range body.Transactions {
		sem <- struct{}{}

		wg.Add(1)

		go func(index int, tx TransactionImportItem) {
			defer func() {
				<-sem
				wg.Done()
			}()

			stopMutex.Lock()
			shouldStop := stopProcessing
			stopMutex.Unlock()
			if shouldStop {
				return
			}

			defer utils.RecoveryWithCallback(&wg, func(r any) {
				errs <- errorInfo{index: index, err: fmt.Errorf("panic recovered: %v", r)}
			})

			transaction, err := c.convertImportedTransaction(&tx, workspaceId, userObjectID)
			if err != nil {
				errs <- errorInfo{index: index, err: err}

				errMsg := err.Error()
				if strings.Contains(errMsg, "membro não encontrado") ||
					strings.Contains(errMsg, "member not found") {
					errMutex.Lock()
					errTypeCount["member_not_found"]++

					if errTypeCount["member_not_found"] > len(body.Transactions)/2 {
						stopMutex.Lock()
						stopProcessing = true
						stopMutex.Unlock()
						select {
						case earlyTerminate <- true:
						default:
						}
					}
					errMutex.Unlock()
				}
				return
			}

			createdTransactions[index] = transaction

			recipeTx := *transaction
			recipeTx.Type = "RECIPE"
			recipeNetBalance := infraHelpers.CalculateOneTransactionBalance(&recipeTx)
			recipeTx.Balance.NetBalance = recipeNetBalance

			importedTransactions[index] = transaction
		}(i, txImport)
	}

	done := make(chan struct{})
	go func() {
		defer utils.Recovery(&wg)
		wg.Wait()
		close(errs)
		close(done)
	}()

	select {
	case <-done:

	case <-earlyTerminate:

	}

	for e := range errs {

		errorMessage := c.translateErrorMessage(e.err.Error())
		validationErrors = append(validationErrors, map[string]any{
			"line":  e.index + 2,
			"error": errorMessage,
		})
	}

	if len(validationErrors) > 0 {
		totalErrors := len(validationErrors)
		displayErrors := validationErrors
		if totalErrors > 50 {
			displayErrors = validationErrors[0:50]
		}

		return helpers.CreateResponse(map[string]any{
			"total":  totalErrors,
			"errors": displayErrors,
		}, http.StatusBadRequest)
	}

	finalTransactions := make([]*models.Transaction, 0, len(importedTransactions))
	for _, tx := range importedTransactions {
		if tx != nil {
			finalTransactions = append(finalTransactions, tx)
		}
	}

	_, err = c.CreateTransactionRepository.CreateMany(finalTransactions)

	if err != nil {
		validationErrors = append(validationErrors, map[string]any{
			"line":  0,
			"error": "Erro ao criar transações: " + err.Error(),
		})
		return helpers.CreateResponse(map[string]any{
			"errors": validationErrors,
		}, http.StatusBadRequest)
	}

	return helpers.CreateResponse(nil, http.StatusCreated)
}

func (c *ImportTransactionController) convertImportedTransaction(txImport *TransactionImportItem, workspaceId, userID primitive.ObjectID) (*models.Transaction, error) {
	parseDate := func(date string) (time.Time, error) {
		location := time.UTC
		return time.ParseInLocation("2006-01-02T15:04:05Z", date, location)
	}

	dueDate, err := parseDate(txImport.DueDate)
	if err != nil {
		return nil, err
	}

	registrationDate, err := parseDate(txImport.RegistrationDate)
	if err != nil {
		return nil, err
	}

	var confirmationDate *time.Time
	if txImport.IsConfirmed && txImport.ConfirmationDate != nil {
		parsedConfDate, err := parseDate(*txImport.ConfirmationDate)
		if err != nil {
			return nil, err
		}
		confirmationDate = &parsedConfDate
	}

	member, err := c.memberCache.getByEmail(txImport.AssignedTo, workspaceId, c.FindMemberByEmailRepository.FindByEmailAndWorkspaceId)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar membro com o email %s: %w", txImport.AssignedTo, err)
	}
	if member == nil {
		return nil, fmt.Errorf("member not found with email: %s", txImport.AssignedTo)
	}

	if strings.TrimSpace(txImport.Account) == "" {
		return nil, errors.New("nome da conta não pode estar vazio")
	}

	account, err := c.accountCache.getOrCreate(txImport.Account, workspaceId, c.FindAccountByNameRepository.FindByNameAndWorkspaceId, c.CreateAccountRepository.Create)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar conta '%s': %w", txImport.Account, err)
	}

	if account == nil {

		c.accountCache.mu.Lock()

		key := cacheKey{name: strings.ToLower(txImport.Account), workspaceId: workspaceId}
		if existingAccount, ok := c.accountCache.items[key]; ok {
			c.accountCache.mu.Unlock()
			account = existingAccount
		} else {

			bank, err := c.bankCache.getByName("Outro", c.FindBankByNameRepository.FindByName)
			if err != nil {
				c.accountCache.mu.Unlock()
				return nil, fmt.Errorf("erro ao buscar banco 'Outro': %w", err)
			}

			if bank == nil {
				c.accountCache.mu.Unlock()
				return nil, errors.New("banco 'Outro' não encontrado no sistema")
			}

			newAccount := &models.Account{
				Id:          primitive.NewObjectID(),
				Name:        txImport.Account,
				Balance:     0,
				WorkspaceId: workspaceId,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
				BankId:      bank.Id,
			}

			account, err = c.CreateAccountRepository.Create(newAccount)
			if err != nil {
				c.accountCache.mu.Unlock()
				return nil, fmt.Errorf("erro ao criar conta '%s': %w", txImport.Account, err)
			}

			c.accountCache.items[key] = account
			c.accountCache.mu.Unlock()
		}
	}

	var categoryId *primitive.ObjectID
	var subCategoryId *primitive.ObjectID

	if txImport.Category != nil && *txImport.Category != "" {

		category, err := c.categoryCache.getOrCreate(*txImport.Category, txImport.Type, workspaceId, c.FindCategoryByNameAndTypeRepository.Find, c.CreateCategoryRepository.Create)
		if err != nil {
			return nil, err
		}

		if category == nil {

			c.categoryCache.mu.Lock()

			key := cacheKey{name: strings.ToLower(*txImport.Category), typ: txImport.Type, workspaceId: workspaceId}
			if category, ok := c.categoryCache.items[key]; ok {
				categoryId = &category.Id
				c.categoryCache.mu.Unlock()
			} else {

				newCategory := &models.Category{
					Id:            primitive.NewObjectID(),
					Name:          *txImport.Category,
					Type:          txImport.Type,
					WorkspaceId:   workspaceId,
					SubCategories: []models.SubCategoryCategory{},
					CreatedAt:     time.Now(),
					UpdatedAt:     time.Now(),
					Icon:          "BookCopy",
				}

				if txImport.SubCategory != nil && *txImport.SubCategory != "" {
					subCatId := primitive.NewObjectID()
					newCategory.SubCategories = append(newCategory.SubCategories, models.SubCategoryCategory{
						Id:   subCatId,
						Name: *txImport.SubCategory,
						Icon: "BookCopy",
					})
					subCategoryId = &subCatId
				}

				category, err = c.CreateCategoryRepository.Create(newCategory)
				if err != nil {
					c.categoryCache.mu.Unlock()
					return nil, fmt.Errorf("error creating category: %w", err)
				}

				c.categoryCache.items[key] = category
				c.categoryCache.mu.Unlock()

				categoryId = &category.Id
			}
		} else {
			categoryId = &category.Id

			if txImport.SubCategory != nil && *txImport.SubCategory != "" {
				found := false
				for _, subCat := range category.SubCategories {
					if strings.EqualFold(subCat.Name, *txImport.SubCategory) {
						subCategoryId = &subCat.Id
						found = true
						break
					}
				}

				if !found {

					updatedCategory := *category
					subCatId := primitive.NewObjectID()
					updatedCategory.SubCategories = append(updatedCategory.SubCategories, models.SubCategoryCategory{
						Id:   subCatId,
						Name: *txImport.SubCategory,
						Icon: "BookCopy",
					})

					updatedCat, err := c.CreateCategoryRepository.Create(&updatedCategory)
					if err != nil {
						return nil, fmt.Errorf("error updating category with new subcategory: %w", err)
					}

					c.categoryCache.mu.Lock()
					c.categoryCache.items[cacheKey{name: strings.ToLower(*txImport.Category), typ: txImport.Type, workspaceId: workspaceId}] = updatedCat
					c.categoryCache.mu.Unlock()

					subCategoryId = &subCatId
				}
			}
		}
	}

	customFields := make([]models.TransactionCustomField, 0)
	for _, cf := range txImport.CustomFields {

		customField, err := c.customFieldCache.getByName(cf.CustomField, workspaceId, c.FindCustomFieldByNameRepository.FindByNameAndWorkspaceId)
		if err != nil {
			return nil, err
		}
		if customField == nil {
			return nil, errors.New("custom field not found: " + cf.CustomField)
		}

		if (customField.TransactionType != txImport.Type) && (customField.TransactionType != "ALL") {
			return nil, errors.New("custom field type mismatch: " + customField.TransactionType + " != " + txImport.Type)
		}

		customFieldId, err := primitive.ObjectIDFromHex(customField.Id)
		if err != nil {
			return nil, errors.New("invalid custom field ID")
		}

		customFields = append(customFields, models.TransactionCustomField{
			CustomFieldId: customFieldId,
			Value:         cf.Value,
			Type:          customField.Type,
		})
	}

	tags := make([]models.TransactionTags, 0)
	for _, tag := range txImport.Tags {
		if tag.Tag == "" {
			continue
		}

		category, err := c.categoryCache.getOrCreate(tag.Tag, "TAG", workspaceId, c.FindCategoryByNameAndTypeRepository.Find, c.CreateCategoryRepository.Create)
		if err != nil {
			return nil, err
		}

		if category == nil {

			newTag := &models.Category{
				Id:            primitive.NewObjectID(),
				Name:          tag.Tag,
				Type:          "TAG",
				Icon:          "BookCopy",
				WorkspaceId:   workspaceId,
				SubCategories: []models.SubCategoryCategory{},
				CreatedAt:     time.Now(),
				UpdatedAt:     time.Now(),
			}

			if tag.SubTag != "" {
				subTagId := primitive.NewObjectID()
				newTag.SubCategories = append(newTag.SubCategories, models.SubCategoryCategory{
					Id:   subTagId,
					Name: tag.SubTag,
					Icon: "BookCopy",
				})

				category, err = c.CreateCategoryRepository.Create(newTag)
				if err != nil {
					return nil, fmt.Errorf("error creating tag: %w", err)
				}

				c.categoryCache.mu.Lock()
				c.categoryCache.items[cacheKey{name: strings.ToLower(tag.Tag), typ: "TAG", workspaceId: workspaceId}] = category
				c.categoryCache.mu.Unlock()

				tags = append(tags, models.TransactionTags{
					TagId:    category.Id,
					SubTagId: subTagId,
				})
			} else {
				category, err = c.CreateCategoryRepository.Create(newTag)
				if err != nil {
					return nil, fmt.Errorf("error creating tag: %w", err)
				}

				c.categoryCache.mu.Lock()
				c.categoryCache.items[cacheKey{name: strings.ToLower(tag.Tag), typ: "TAG", workspaceId: workspaceId}] = category
				c.categoryCache.mu.Unlock()

				tags = append(tags, models.TransactionTags{
					TagId:    category.Id,
					SubTagId: primitive.NilObjectID,
				})
			}
		} else {
			if !strings.EqualFold(category.Type, "TAG") {
				return nil, errors.New("category is not a tag: " + tag.Tag)
			}

			tags = append(tags, models.TransactionTags{
				TagId:    category.Id,
				SubTagId: primitive.NilObjectID,
			})

			if tag.SubTag == "" {
				continue
			}

			found := false
			for _, subCat := range category.SubCategories {
				if strings.EqualFold(subCat.Name, tag.SubTag) {
					tags[len(tags)-1].SubTagId = subCat.Id
					found = true
					break
				}
			}

			if found {
				continue
			}

			updatedTag := *category
			subTagId := primitive.NewObjectID()
			updatedTag.SubCategories = append(updatedTag.SubCategories, models.SubCategoryCategory{
				Id:   subTagId,
				Name: tag.SubTag,
			})

			updatedTagPtr, err := c.CreateCategoryRepository.Create(&updatedTag)
			if err != nil {
				return nil, fmt.Errorf("error updating tag with new subtag: %w", err)
			}

			c.categoryCache.mu.Lock()
			c.categoryCache.items[cacheKey{name: strings.ToLower(tag.Tag), typ: "TAG", workspaceId: workspaceId}] = updatedTagPtr
			c.categoryCache.mu.Unlock()

			tags[len(tags)-1].SubTagId = subTagId
		}
	}

	var repeatSettings *models.TransactionRepeatSettings
	if txImport.Frequency == "REPEAT" {
		repeatSettings = &models.TransactionRepeatSettings{
			InitialInstallment: txImport.RepeatSettings.InitialInstallment,
			Count:              txImport.RepeatSettings.Count,
			Interval:           txImport.RepeatSettings.Interval,
			CustomDay:          txImport.RepeatSettings.CustomDay,
		}

		if txImport.RepeatSettings.Interval == "CUSTOM" {
			repeatSettings.CustomDay = txImport.RepeatSettings.CustomDay
		}
	}

	transaction := &models.Transaction{
		Id:          primitive.NewObjectID(),
		Name:        txImport.Name,
		Description: txImport.Description,
		Invoice:     txImport.Invoice,
		Type:        txImport.Type,
		Supplier:    txImport.Supplier,
		AssignedTo:  member.MemberId,
		CreatedBy:   userID,
		WorkspaceId: workspaceId,
		Balance: models.TransactionBalance{
			Value:              txImport.Balance.Value,
			Discount:           txImport.Balance.Discount,
			Interest:           txImport.Balance.Interest,
			DiscountPercentage: txImport.Balance.DiscountPercentage,
			InterestPercentage: txImport.Balance.InterestPercentage,
		},
		Frequency:        txImport.Frequency,
		RepeatSettings:   repeatSettings,
		DueDate:          dueDate,
		IsConfirmed:      txImport.IsConfirmed,
		CategoryId:       categoryId,
		SubCategoryId:    subCategoryId,
		Tags:             tags,
		AccountId:        &account.Id,
		RegistrationDate: registrationDate,
		ConfirmationDate: confirmationDate,
		CustomFields:     customFields,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	return transaction, nil
}

type cacheKey struct {
	name        string
	typ         string
	workspaceId primitive.ObjectID
}

type categoryCache struct {
	mu    sync.RWMutex
	items map[cacheKey]*models.Category
}

type accountCache struct {
	mu    sync.RWMutex
	items map[cacheKey]*models.Account
}

type memberCache struct {
	mu    sync.RWMutex
	items map[cacheKey]*models.Member
}

type customFieldCache struct {
	mu    sync.RWMutex
	items map[cacheKey]*models.CustomField
}

type bankCache struct {
	mu    sync.RWMutex
	items map[string]*models.Bank
}

func (c *categoryCache) getOrCreate(name, typ string, workspaceId primitive.ObjectID, findFn func(string, string, primitive.ObjectID) (*models.Category, error), createFn func(*models.Category) (*models.Category, error)) (*models.Category, error) {
	key := cacheKey{name: strings.ToLower(name), typ: typ, workspaceId: workspaceId}

	c.mu.RLock()
	category, ok := c.items[key]
	c.mu.RUnlock()

	if ok {
		return category, nil
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if category, ok := c.items[key]; ok {
		return category, nil
	}

	category, err := findFn(name, typ, workspaceId)
	if err != nil {
		return nil, err
	}

	if category != nil {
		c.items[key] = category
		return category, nil
	}

	return nil, nil
}

func (c *accountCache) getOrCreate(name string, workspaceId primitive.ObjectID, findFn func(string, primitive.ObjectID) (*models.Account, error), createFn func(*models.Account) (*models.Account, error)) (*models.Account, error) {
	key := cacheKey{name: strings.ToLower(name), workspaceId: workspaceId}

	c.mu.RLock()
	account, ok := c.items[key]
	c.mu.RUnlock()

	if ok {
		return account, nil
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if account, ok := c.items[key]; ok {
		return account, nil
	}

	account, err := findFn(name, workspaceId)
	if err != nil {
		return nil, err
	}

	if account != nil {
		c.items[key] = account
		return account, nil
	}

	return nil, nil
}

func (c *memberCache) getByEmail(email string, workspaceId primitive.ObjectID, findFn func(string, primitive.ObjectID) (*models.Member, error)) (*models.Member, error) {
	key := cacheKey{name: strings.ToLower(email), workspaceId: workspaceId}

	c.mu.RLock()
	member, ok := c.items[key]
	c.mu.RUnlock()

	if ok {
		return member, nil
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if member, ok := c.items[key]; ok {
		return member, nil
	}

	member, err := findFn(email, workspaceId)
	if err != nil {
		return nil, err
	}

	if member != nil {
		c.items[key] = member
	}

	return member, nil
}

func (c *customFieldCache) getByName(name string, workspaceId primitive.ObjectID, findFn func(string, primitive.ObjectID) (*models.CustomField, error)) (*models.CustomField, error) {
	key := cacheKey{name: strings.ToLower(name), workspaceId: workspaceId}

	c.mu.RLock()
	customField, ok := c.items[key]
	c.mu.RUnlock()

	if ok {
		return customField, nil
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if customField, ok := c.items[key]; ok {
		return customField, nil
	}

	customField, err := findFn(name, workspaceId)
	if err != nil {
		return nil, err
	}

	if customField != nil {
		c.items[key] = customField
	}

	return customField, nil
}

func (c *bankCache) getByName(name string, findFn func(string) (*models.Bank, error)) (*models.Bank, error) {

	c.mu.RLock()
	bank, ok := c.items[strings.ToLower(name)]
	c.mu.RUnlock()

	if ok {
		return bank, nil
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if bank, ok := c.items[strings.ToLower(name)]; ok {
		return bank, nil
	}

	bank, err := findFn(name)
	if err != nil {
		return nil, err
	}

	if bank != nil {
		c.items[strings.ToLower(name)] = bank
	}

	return bank, nil
}

type ColumnDef struct {
	Key           string `json:"key"`
	KeyToMap      string `json:"keyToMap"`
	IsCustomField bool   `json:"isCustomField"`
}

func (c *ImportTransactionController) ParseMultipartAndMap(r *http.Request) ([]TransactionImportItem, error) {

	if err := r.ParseMultipartForm(32 << 20); err != nil {
		return nil, fmt.Errorf("invalid multipart form: %w", err)
	}

	columnsJSON := r.FormValue("columns")
	if columnsJSON == "" {
		return nil, fmt.Errorf("campo 'columns' vazio ou ausente no form-data")
	}

	var columns []ColumnDef
	if err := json.Unmarshal([]byte(columnsJSON), &columns); err != nil {
		return nil, fmt.Errorf("JSON de colunas inválido: %w", err)
	}

	if len(columns) == 0 {
		return nil, fmt.Errorf("nenhuma coluna definida no mapeamento")
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		return nil, fmt.Errorf("missing 'file' field in form-data: %w", err)
	}
	defer file.Close()

	var rawRows []map[string]any
	ext := strings.ToLower(filepath.Ext(header.Filename))
	switch ext {
	case ".csv":
		rawRows, err = ParseCSV(file)
	case ".xlsx", ".xlsm", ".xls":
		rawRows, err = ParseXLSX(file)
	default:
		return nil, fmt.Errorf("unsupported file type %s", ext)
	}
	if err != nil {
		return nil, err
	}

	mappedRows := ApplyMapping(rawRows, columns)

	txs := make([]TransactionImportItem, 0, len(mappedRows))
	for _, row := range mappedRows {

		b, _ := json.Marshal(row)
		var tx TransactionImportItem
		if err := json.Unmarshal(b, &tx); err != nil {
			return nil, fmt.Errorf("row to struct error: %w", err)
		}
		txs = append(txs, tx)
	}
	return txs, nil
}

func ParseCSV(r io.Reader) ([]map[string]any, error) {
	cr := csv.NewReader(r)
	cr.TrimLeadingSpace = true
	cr.LazyQuotes = true
	headers, err := cr.Read()
	if err != nil {
		return nil, fmt.Errorf("csv header: %w", err)
	}

	for i, h := range headers {

		headers[i] = strings.ReplaceAll(h, "\"", "")
	}

	var rows []map[string]any
	for {
		rec, err := cr.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("csv row: %w", err)
		}

		if isEmptyRow(rec) {
			continue
		}

		row := make(map[string]any)
		for i, h := range headers {
			row[h] = rec[i]
		}
		rows = append(rows, row)
	}
	return rows, nil
}

func ParseXLSX(r multipart.File) ([]map[string]any, error) {

	buf := bytes.NewBuffer(nil)
	if _, err := io.Copy(buf, r); err != nil {
		return nil, fmt.Errorf("copy xlsx: %w", err)
	}
	f, err := excelize.OpenReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		return nil, fmt.Errorf("open xlsx: %w", err)
	}
	defer f.Close()

	sheet := f.GetSheetName(0)
	rowsIter, err := f.Rows(sheet)
	if err != nil {
		return nil, err
	}

	if !rowsIter.Next() {
		return nil, fmt.Errorf("xlsx empty sheet")
	}
	headers, _ := rowsIter.Columns()
	var rows []map[string]any
	for rowsIter.Next() {
		cols, _ := rowsIter.Columns()

		if isEmptyRow(cols) {
			continue
		}

		row := make(map[string]any)
		for i, h := range headers {
			if i < len(cols) {
				row[h] = cols[i]
			} else {
				row[h] = ""
			}
		}
		rows = append(rows, row)
	}
	return rows, nil
}

func isEmptyRow(cols []string) bool {
	if len(cols) == 0 {
		return true
	}

	for _, col := range cols {

		if strings.TrimSpace(col) != "" {
			return false
		}
	}

	return true
}

func ApplyMapping(rows []map[string]any, defs []ColumnDef) []map[string]any {
	if len(defs) == 0 {
		return rows
	}

	headerMap := make(map[string]string)
	if len(rows) > 0 {
		for k := range rows[0] {
			normalizedKey := normalize(k)
			headerMap[normalizedKey] = k
		}
	}

	numericFields := map[string]bool{
		"balance.value":              true,
		"balance.discount":           true,
		"balance.interest":           true,
		"balance.discountPercentage": true,
		"balance.interestPercentage": true,
	}

	mapped := make([]map[string]any, len(rows))
	for i, row := range rows {
		mappedToAppend := make(map[string]any)

		for k, v := range row {
			mappedToAppend[k] = v
		}

		for _, col := range defs {
			normalizedKeyToMap := normalize(col.KeyToMap)

			if col.Key == "" || col.KeyToMap == "" {
				continue
			}

			if col.IsCustomField {

				id := col.Key
				value := row[strings.TrimSpace(col.KeyToMap)]

				cfSlice, ok := mappedToAppend["customFields"].([]map[string]any)
				if !ok {
					cfSlice = []map[string]any{}
				}

				cfSlice = append(cfSlice, map[string]any{
					"customField": id,
					"value":       value,
				})
				mappedToAppend["customFields"] = cfSlice
				continue
			}

			if col.Key == "tags" {
				tagValue, exists := row[strings.TrimSpace(col.KeyToMap)]
				if !exists || tagValue == nil {
					continue
				}

				tags, ok := tagValue.(string)
				if !ok {
					// If not a string, skip this tag
					continue
				}

				tagsSlice, ok := mappedToAppend["tags"].([]map[string]any)
				if !ok {
					tagsSlice = []map[string]any{}
				}

				tagSplited := strings.Split(tags, ",")
				for _, tag := range tagSplited {
					if strings.TrimSpace(tag) == "" {
						continue
					}
					splitSubTagAndTag := strings.Split(tag, "-")
					if len(splitSubTagAndTag) == 2 {
						tagsSlice = append(tagsSlice, map[string]any{
							"tag":    strings.TrimSpace(splitSubTagAndTag[0]),
							"subTag": strings.TrimSpace(splitSubTagAndTag[1]),
						})
					} else {
						continue
					}
				}

				mappedToAppend["tags"] = tagsSlice
				continue
			}

			if strings.Contains(col.Key, ".") {
				parts := strings.SplitN(col.Key, ".", 2)
				parentKey := parts[0]
				childKey := parts[1]

				var foundValue any
				var found bool

				if originalKey, ok := headerMap[normalizedKeyToMap]; ok {
					if val, ok := row[originalKey]; ok {
						foundValue = val
						found = true
					}
				}

				if !found {
					for k, v := range row {
						normalizedK := normalize(k)
						if normalizedK == normalizedKeyToMap {
							foundValue = v
							found = true
							break
						}
					}
				}

				if found {

					if strVal, isStr := foundValue.(string); isStr && numericFields[col.Key] {
						cleanVal := strings.ReplaceAll(strVal, ",", ".")

						cleanVal = strings.TrimSpace(cleanVal)

						if cleanVal == "" {
							cleanVal = "0"
						}

						floatVal, err := strconv.ParseFloat(cleanVal, 64)
						if err != nil {
							fmt.Println(err)
							continue
						}

						foundValue = floatVal
						fmt.Println(foundValue)
					}

					parentObj, ok := mappedToAppend[parentKey].(map[string]any)
					if !ok {

						parentObj = make(map[string]any)
						mappedToAppend[parentKey] = parentObj
					}

					parentObj[childKey] = foundValue
				}
				continue
			}

			if originalKey, ok := headerMap[normalizedKeyToMap]; ok {
				if val, ok := row[originalKey]; ok {
					mappedToAppend[col.Key] = val
					if col.Key != originalKey {
						delete(mappedToAppend, originalKey)
					}
					continue
				}
			}

			found := false
			var foundKey string
			var foundValue any

			for k, v := range row {

				normalizedK := normalize(k)

				if normalizedK == normalizedKeyToMap {
					found = true
					foundKey = k
					foundValue = v
					break
				}
			}

			if found {
				mappedToAppend[col.Key] = foundValue
				if col.Key != foundKey {
					delete(mappedToAppend, foundKey)
				}
			}
		}
		mapped[i] = mappedToAppend
	}
	return mapped
}

func normalize(s string) string {

	s = strings.ToLower(s)

	s = strings.TrimSpace(s)

	var result []rune
	for _, r := range s {
		if r > 32 && r < 127 {
			result = append(result, r)
		}
	}

	return string(result)
}

func (c *ImportTransactionController) ParseAllDatesAndTypes(transactions []TransactionImportItem) ([]TransactionImportItem, error) {
	var dateErrors []string

	for i := range transactions {
		transactions[i].Frequency = "DO_NOT_REPEAT"

		if strings.ToLower(transactions[i].Type) == "receita" || strings.ToLower(transactions[i].Type) == "recurring" {
			transactions[i].Type = "RECIPE"
		}

		if strings.ToLower(transactions[i].Type) == "despesa" || strings.ToLower(transactions[i].Type) == "expense" {
			transactions[i].Type = "EXPENSE"
		}

		if *transactions[i].ConfirmationDate == "" {
			transactions[i].ConfirmationDate = nil
			transactions[i].IsConfirmed = false
		}

		transactions[i].DueDate = strings.ReplaceAll(transactions[i].DueDate, "-", "/")
		transactions[i].RegistrationDate = strings.ReplaceAll(transactions[i].RegistrationDate, "-", "/")
		if transactions[i].ConfirmationDate != nil {
			*transactions[i].ConfirmationDate = strings.ReplaceAll(*transactions[i].ConfirmationDate, "-", "/")
		}

		if transactions[i].DueDate != "" {
			parts := strings.Split(transactions[i].DueDate, "/")
			if len(parts) == 3 {
				for j := 0; j < 2; j++ {
					if len(parts[j]) == 1 {
						parts[j] = "0" + parts[j]
					}
				}
				transactions[i].DueDate = strings.Join(parts, "/")
			}

			// Tenta formatos DD/MM/AAAA, DD/MM/AA, MM/DD/AAAA e MM/DD/AA
			t, err := time.Parse("02/01/2006", transactions[i].DueDate)
			if err != nil {
				t, err = time.Parse("02/01/06", transactions[i].DueDate)
				if err != nil {
					// Tenta formato MM/DD/AAAA
					t, err = time.Parse("01/02/2006", transactions[i].DueDate)
					if err != nil {
						// Tenta formato MM/DD/AA
						t, err = time.Parse("01/02/06", transactions[i].DueDate)
						if err != nil {
							dateErrors = append(dateErrors, fmt.Sprintf("A data de vencimento "+transactions[i].DueDate+" na linha %d não está em um formato válido. Formatos aceitos: DD/MM/AAAA, DD/MM/AA, MM/DD/AAAA ou MM/DD/AA", i+2))
							continue
						}
					}
				}
			}
			transactions[i].DueDate = t.UTC().Format("2006-01-02T15:04:05Z")
		}

		if transactions[i].RegistrationDate != "" {
			parts := strings.Split(transactions[i].RegistrationDate, "/")
			if len(parts) == 3 {
				for j := 0; j < 2; j++ {
					if len(parts[j]) == 1 {
						parts[j] = "0" + parts[j]
					}
				}
				transactions[i].RegistrationDate = strings.Join(parts, "/")
			}

			// Tenta formatos DD/MM/AAAA, DD/MM/AA, MM/DD/AAAA e MM/DD/AA
			t, err := time.Parse("02/01/2006", transactions[i].RegistrationDate)
			if err != nil {
				t, err = time.Parse("02/01/06", transactions[i].RegistrationDate)
				if err != nil {
					// Tenta formato MM/DD/AAAA
					t, err = time.Parse("01/02/2006", transactions[i].RegistrationDate)
					if err != nil {
						// Tenta formato MM/DD/AA
						t, err = time.Parse("01/02/06", transactions[i].RegistrationDate)
						if err != nil {
							dateErrors = append(dateErrors, fmt.Sprintf("A data de registro na linha %d não está em um formato válido. Formatos aceitos: DD/MM/AAAA, DD/MM/AA, MM/DD/AAAA ou MM/DD/AA", i+2))
							continue
						}
					}
				}
			}
			transactions[i].RegistrationDate = t.UTC().Format("2006-01-02T15:04:05Z")
		}

		if transactions[i].ConfirmationDate != nil && *transactions[i].ConfirmationDate != "" {
			parts := strings.Split(*transactions[i].ConfirmationDate, "/")
			if len(parts) == 3 {
				for j := 0; j < 2; j++ {
					if len(parts[j]) == 1 {
						parts[j] = "0" + parts[j]
					}
				}
				*transactions[i].ConfirmationDate = strings.Join(parts, "/")
			}

			// Tenta formatos DD/MM/AAAA, DD/MM/AA, MM/DD/AAAA e MM/DD/AA
			t, err := time.Parse("02/01/2006", *transactions[i].ConfirmationDate)
			if err != nil {
				t, err = time.Parse("02/01/06", *transactions[i].ConfirmationDate)
				if err != nil {
					// Tenta formato MM/DD/AAAA
					t, err = time.Parse("01/02/2006", *transactions[i].ConfirmationDate)
					if err != nil {
						// Tenta formato MM/DD/AA
						t, err = time.Parse("01/02/06", *transactions[i].ConfirmationDate)
						if err != nil {
							dateErrors = append(dateErrors, fmt.Sprintf("A data de confirmação na linha %d não está em um formato válido. Formatos aceitos: DD/MM/AAAA, DD/MM/AA, MM/DD/AAAA ou MM/DD/AA", i+2))
							continue
						}
					}
				}
			}
			formattedDate := t.UTC().Format("2006-01-02T15:04:05Z")
			transactions[i].ConfirmationDate = &formattedDate
			transactions[i].IsConfirmed = true
		}
	}

	if len(dateErrors) > 0 {
		return transactions, fmt.Errorf("erros na análise de datas: %s", strings.Join(dateErrors, "; "))
	}

	return transactions, nil
}

func (c *ImportTransactionController) translateErrorMessage(errorMsg string) string {

	errorTranslations := map[string]string{
		"member not found with email":                  "Usuário não encontrado. Verifique se o email está cadastrado no sistema",
		"bank not found":                               "Banco não encontrado no sistema",
		"error creating account":                       "Erro ao criar conta. Por favor, tente novamente",
		"nome da conta não pode estar vazio":           "O nome da conta é obrigatório",
		"banco 'Outro' não encontrado no sistema":      "Banco padrão 'Outro' não encontrado no sistema",
		"erro ao criar conta":                          "Erro ao criar conta. Por favor, tente novamente",
		"erro ao buscar conta":                         "Erro ao buscar conta. Verifique se a conta existe no sistema",
		"erro ao buscar banco":                         "Erro ao buscar banco. Verifique se o banco existe no sistema",
		"error creating category":                      "Erro ao criar categoria. Por favor, tente novamente",
		"error updating category with new subcategory": "Erro ao adicionar subcategoria. Por favor, tente novamente",
		"custom field not found":                       "Campo personalizado não encontrado no sistema",
		"custom field type mismatch":                   "O tipo do campo personalizado não é compatível com o tipo da transação",
		"invalid custom field ID":                      "ID do campo personalizado inválido",
		"category is not a tag":                        "A categoria selecionada não é uma tag",
		"error creating tag":                           "Erro ao criar tag. Por favor, tente novamente",
		"error updating tag with new subtag":           "Erro ao atualizar tag com nova subtag",
		"panic recovered":                              "Ocorreu um erro inesperado. Por favor, tente novamente",
		"invalid user ID format":                       "Formato de ID de usuário inválido",
		"invalid workspace ID format":                  "Formato de ID de espaço de trabalho inválido",
	}

	for engMsg, ptMsg := range errorTranslations {
		if strings.Contains(errorMsg, engMsg) {
			return strings.Replace(errorMsg, engMsg, ptMsg, 1)
		}
	}

	// Return the error message as is for our new date format errors
	// They already have line numbers and clear messages
	if strings.Contains(errorMsg, "A data de confirmação na linha") ||
		strings.Contains(errorMsg, "A data de vencimento na linha") ||
		strings.Contains(errorMsg, "A data de registro na linha") {
		return errorMsg
	}

	return errorMsg
}

func (c *ImportTransactionController) translateValidationError(err validator.FieldError) string {
	fieldName := err.Field()
	tag := err.Tag()
	param := err.Param()

	if strings.Contains(fieldName, "Transactions") && strings.Contains(fieldName, ".") {
		parts := strings.Split(fieldName, ".")
		if len(parts) > 1 {
			fieldName = parts[len(parts)-1]
		}
	}

	fieldTranslations := map[string]string{
		"Name":             "Nome",
		"Description":      "Descrição",
		"Invoice":          "Fatura",
		"Type":             "Tipo",
		"Supplier":         "Fornecedor",
		"AssignedTo":       "Responsável",
		"Balance":          "Balanço",
		"Value":            "Valor",
		"Discount":         "Desconto",
		"Interest":         "Juros",
		"DueDate":          "Data de vencimento",
		"Account":          "Conta",
		"Category":         "Categoria",
		"SubCategory":      "Subcategoria",
		"ConfirmationDate": "Data de confirmação",
		"RegistrationDate": "Data de registro",
	}

	fieldTranslated := fieldName
	if translated, ok := fieldTranslations[fieldName]; ok {
		fieldTranslated = translated
	}

	if param == "EXPENSE RECIPE" {
		param = "Despesa Receita"
	}

	switch tag {
	case "required":
		return fieldTranslated + " é obrigatório"
	case "min":
		return fieldTranslated + " deve ter no mínimo " + param + " caracteres"
	case "max":
		return fieldTranslated + " deve ter no máximo " + param + " caracteres"
	case "email":
		return fieldTranslated + " deve ser um email válido"
	case "oneof":
		return fieldTranslated + " deve ser um dos seguintes valores: " + param
	case "datetime":
		return fieldTranslated + " deve estar no formato de data válido (DD/MM/AAAA)"
	case "excluded_if":
		if fieldName == "ConfirmationDate" {
			return "Data de confirmação só deve ser informada para transações confirmadas"
		}
		return fieldTranslated + " não deve ser informado nas condições atuais"
	case "required_if":
		if fieldName == "ConfirmationDate" {
			return "Data de confirmação é obrigatória para transações confirmadas"
		}
		return fieldTranslated + " é obrigatório nas condições atuais"
	default:
		return "Erro de validação no campo " + fieldTranslated + ": " + tag
	}
}
