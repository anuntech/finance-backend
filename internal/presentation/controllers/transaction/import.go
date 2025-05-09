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
	"strconv"
	"strings"
	"sync"
	"time"

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

	// Caches
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
	Category    *string `json:"category" validate:"omitempty"`
	SubCategory *string `json:"subCategory" validate:"omitempty"`
	Tags        []struct {
		Tag    string `json:"tag" validate:"omitempty"`
		SubTag string `json:"subTag" validate:"omitempty"`
	} `json:"tags" validate:"omitempty"`
	CustomFields []struct {
		CustomField string `json:"customField" validate:"required"`
		Value       string `json:"value" validate:"required,max=100"`
	} `json:"customFields" validate:"omitempty"`
	Account          string  `json:"account"`
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

	transactions, err = c.ParseAllDatesAndTypes(transactions)
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "error parsing all dates: " + err.Error(),
		}, http.StatusBadRequest)
	}

	body := &ImportTransactionBody{
		Transactions: transactions,
	}

	// bodyJSON, _ := json.MarshalIndent(body, "", "  ")
	// fmt.Println(string(bodyJSON))
	validationErrors := []map[string]any{}

	if err := c.Validate.Struct(body); err != nil {
		// Converter os erros de validação para o mesmo formato de array

		if errs, ok := err.(validator.ValidationErrors); ok {
			for _, e := range errs {
				// Encontrar o índice da transação com erro
				field := e.Field()
				index := 0

				// Extrair índice da transação se o erro for em um item do array
				if strings.HasPrefix(field, "Transactions[") {
					parts := strings.Split(field, "[")
					if len(parts) > 1 {
						indexPart := strings.Split(parts[1], "]")[0]
						if idx, err := strconv.Atoi(indexPart); err == nil {
							index = idx
						}
					}
				}

				// Traduzir a mensagem de erro
				errorMsg := c.translateValidationError(e)
				validationErrors = append(validationErrors, map[string]any{
					"linha": index + 2,
					"erro":  errorMsg,
				})
			}
		} else {
			// Caso não seja um erro de validação padrão
			validationErrors = append(validationErrors, map[string]any{
				"linha": 0,
				"erro":  "Erro de validação: " + err.Error(),
			})
		}
	}

	userID := r.Header.Get("userId")
	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "invalid user ID format",
		}, http.StatusBadRequest)
	}

	workspaceIdStr := r.Header.Get("workspaceId")
	workspaceId, err := primitive.ObjectIDFromHex(workspaceIdStr)
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "invalid workspace ID format",
		}, http.StatusBadRequest)
	}

	LIMIT := 10000
	if len(body.Transactions) > LIMIT {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "maximum of " + strconv.Itoa(LIMIT) + " transactions per import",
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
	const workers = 100
	sem := make(chan struct{}, workers)

	for i, txImport := range body.Transactions {
		// 1) bloqueia aqui se já houver 100 em execução
		sem <- struct{}{}

		// 2) agora sim a vaga foi conquistada: conta 1 no WaitGroup
		wg.Add(1)

		go func(index int, tx TransactionImportItem) {
			defer func() {
				<-sem     // libera a vaga
				wg.Done() // decrementa só quem realmente executou
			}()

			defer utils.RecoveryWithCallback(&wg, func(r any) {
				errs <- errorInfo{index: index + 1, err: fmt.Errorf("panic recovered: %v", r)}
			})

			transaction, err := c.convertImportedTransaction(&tx, workspaceId, userObjectID)
			if err != nil {
				errs <- errorInfo{index: index + 1, err: err}
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

	go func() {
		defer utils.Recovery(&wg)
		wg.Wait()
		close(errs)
	}()

	// Coletar todos os erros em vez de retornar no primeiro
	for e := range errs {
		// Traduzir a mensagem de erro para português e adicionar ao array
		errorMessage := c.translateErrorMessage(e.err.Error())
		validationErrors = append(validationErrors, map[string]any{
			"linha": e.index + 1,
			"erro":  errorMessage,
		})
	}

	// Se houver erros, retornar todos eles
	if len(validationErrors) > 0 {
		return helpers.CreateResponse(map[string]any{
			"erros": validationErrors,
		}, http.StatusBadRequest)
	}

	finalTransactions := make([]*models.Transaction, 0, len(importedTransactions))
	for _, tx := range importedTransactions {
		if tx != nil {
			finalTransactions = append(finalTransactions, tx)
		}
	}

	createdTransactions, err = c.CreateTransactionRepository.CreateMany(finalTransactions)

	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: fmt.Sprintf("erro ao criar transações: %s", err.Error()),
		}, http.StatusBadRequest)
	}

	return helpers.CreateResponse(createdTransactions, http.StatusCreated)
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

	// Use cache for member lookup
	member, err := c.memberCache.getByEmail(txImport.AssignedTo, workspaceId, c.FindMemberByEmailRepository.FindByEmailAndWorkspaceId)
	if err != nil {
		return nil, err
	}
	if member == nil {
		return nil, errors.New("member not found with email: " + txImport.AssignedTo)
	}

	// Try to find account with cache, create if not found
	account, err := c.accountCache.getOrCreate(txImport.Account, workspaceId, c.FindAccountByNameRepository.FindByNameAndWorkspaceId, c.CreateAccountRepository.Create)
	if err != nil {
		return nil, err
	}

	if account == nil {
		// Use bank cache
		bank, err := c.bankCache.getByName("Outro", c.FindBankByNameRepository.FindByName)
		if err != nil {
			return nil, err
		}

		if bank == nil {
			return nil, errors.New("bank not found: Outro")
		}

		// Create a new account
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
			return nil, fmt.Errorf("error creating account: %w", err)
		}

		// Add to cache
		c.accountCache.mu.Lock()
		c.accountCache.items[cacheKey{name: strings.ToLower(txImport.Account), workspaceId: workspaceId}] = account
		c.accountCache.mu.Unlock()
	}

	var categoryId *primitive.ObjectID
	var subCategoryId *primitive.ObjectID

	if txImport.Category != nil && *txImport.Category != "" {
		// Use category cache
		category, err := c.categoryCache.getOrCreate(*txImport.Category, txImport.Type, workspaceId, c.FindCategoryByNameAndTypeRepository.Find, c.CreateCategoryRepository.Create)
		if err != nil {
			return nil, err
		}

		if category == nil {
			// Create a new category
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

			// If subcategory is specified, add it to the new category
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
				return nil, fmt.Errorf("error creating category: %w", err)
			}

			// Add to cache
			c.categoryCache.mu.Lock()
			c.categoryCache.items[cacheKey{name: strings.ToLower(*txImport.Category), typ: txImport.Type, workspaceId: workspaceId}] = category
			c.categoryCache.mu.Unlock()

			categoryId = &category.Id
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
					// Add the subcategory to the existing category
					updatedCategory := *category
					subCatId := primitive.NewObjectID()
					updatedCategory.SubCategories = append(updatedCategory.SubCategories, models.SubCategoryCategory{
						Id:   subCatId,
						Name: *txImport.SubCategory,
						Icon: "BookCopy",
					})

					// Update the category with the new subcategory
					updatedCat, err := c.CreateCategoryRepository.Create(&updatedCategory)
					if err != nil {
						return nil, fmt.Errorf("error updating category with new subcategory: %w", err)
					}

					// Update cache
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
		// Use custom field cache
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

		// Use tag cache (categories of type TAG)
		category, err := c.categoryCache.getOrCreate(tag.Tag, "TAG", workspaceId, c.FindCategoryByNameAndTypeRepository.Find, c.CreateCategoryRepository.Create)
		if err != nil {
			return nil, err
		}

		if category == nil {
			// Create a new tag category
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

			// If subtag is specified, add it to the new tag
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

				// Add to cache
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

				// Add to cache
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
			// Add the subtag to the existing tag
			updatedTag := *category
			subTagId := primitive.NewObjectID()
			updatedTag.SubCategories = append(updatedTag.SubCategories, models.SubCategoryCategory{
				Id:   subTagId,
				Name: tag.SubTag,
			})

			// Update the tag with the new subtag
			updatedTagPtr, err := c.CreateCategoryRepository.Create(&updatedTag)
			if err != nil {
				return nil, fmt.Errorf("error updating tag with new subtag: %w", err)
			}

			// Update cache
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

// Cache structures
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

// Cache methods
func (c *categoryCache) getOrCreate(name, typ string, workspaceId primitive.ObjectID, findFn func(string, string, primitive.ObjectID) (*models.Category, error), createFn func(*models.Category) (*models.Category, error)) (*models.Category, error) {
	key := cacheKey{name: strings.ToLower(name), typ: typ, workspaceId: workspaceId}

	// Try to get from cache first
	c.mu.RLock()
	category, ok := c.items[key]
	c.mu.RUnlock()
	if ok {
		return category, nil
	}

	// Not in cache, find in database
	category, err := findFn(name, typ, workspaceId)
	if err != nil {
		return nil, err
	}

	// If found, update cache and return
	if category != nil {
		c.mu.Lock()
		c.items[key] = category
		c.mu.Unlock()
		return category, nil
	}

	// Not found, needs to be created
	return nil, nil
}

func (c *accountCache) getOrCreate(name string, workspaceId primitive.ObjectID, findFn func(string, primitive.ObjectID) (*models.Account, error), createFn func(*models.Account) (*models.Account, error)) (*models.Account, error) {
	key := cacheKey{name: strings.ToLower(name), workspaceId: workspaceId}

	// Try to get from cache first
	c.mu.RLock()
	account, ok := c.items[key]
	c.mu.RUnlock()
	if ok {
		return account, nil
	}

	// Not in cache, find in database
	account, err := findFn(name, workspaceId)
	if err != nil {
		return nil, err
	}

	// If found, update cache and return
	if account != nil {
		c.mu.Lock()
		c.items[key] = account
		c.mu.Unlock()
		return account, nil
	}

	// Not found, needs to be created
	return nil, nil
}

func (c *memberCache) getByEmail(email string, workspaceId primitive.ObjectID, findFn func(string, primitive.ObjectID) (*models.Member, error)) (*models.Member, error) {
	key := cacheKey{name: strings.ToLower(email), workspaceId: workspaceId}

	// Try to get from cache first
	c.mu.RLock()
	member, ok := c.items[key]
	c.mu.RUnlock()
	if ok {
		return member, nil
	}

	// Not in cache, find in database
	member, err := findFn(email, workspaceId)
	if err != nil {
		return nil, err
	}

	// If found, update cache and return
	if member != nil {
		c.mu.Lock()
		c.items[key] = member
		c.mu.Unlock()
	}

	return member, nil
}

func (c *customFieldCache) getByName(name string, workspaceId primitive.ObjectID, findFn func(string, primitive.ObjectID) (*models.CustomField, error)) (*models.CustomField, error) {
	key := cacheKey{name: strings.ToLower(name), workspaceId: workspaceId}

	// Try to get from cache first
	c.mu.RLock()
	customField, ok := c.items[key]
	c.mu.RUnlock()
	if ok {
		return customField, nil
	}

	// Not in cache, find in database
	customField, err := findFn(name, workspaceId)
	if err != nil {
		return nil, err
	}

	// If found, update cache and return
	if customField != nil {
		c.mu.Lock()
		c.items[key] = customField
		c.mu.Unlock()
	}

	return customField, nil
}

func (c *bankCache) getByName(name string, findFn func(string) (*models.Bank, error)) (*models.Bank, error) {
	// Try to get from cache first
	c.mu.RLock()
	bank, ok := c.items[strings.ToLower(name)]
	c.mu.RUnlock()
	if ok {
		return bank, nil
	}

	// Not in cache, find in database
	bank, err := findFn(name)
	if err != nil {
		return nil, err
	}

	// If found, update cache and return
	if bank != nil {
		c.mu.Lock()
		c.items[strings.ToLower(name)] = bank
		c.mu.Unlock()
	}

	return bank, nil
}

type ColumnDef struct {
	Key           string `json:"key"`           // novo nome
	KeyToMap      string `json:"keyToMap"`      // chave original
	IsCustomField bool   `json:"isCustomField"` // se é custom field
}

// parseMultipartAndMap lê o formulário, decodifica Columns, faz parse do arquivo
// CSV ou XLSX e devolve o slice de TransactionImportItem já mapeado.
func (c *ImportTransactionController) ParseMultipartAndMap(r *http.Request) ([]TransactionImportItem, error) {
	// ~32 MB de memória antes de cair em arquivo temporário
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		return nil, fmt.Errorf("invalid multipart form: %w", err)
	}

	// --- Columns ----------------------------------------------------------
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

	// --- File ---
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

	// --- Apply mapping ----------------------------------------------------
	mappedRows := ApplyMapping(rawRows, columns)

	// --- Convert to TransactionImportItem ---------------------------------
	txs := make([]TransactionImportItem, 0, len(mappedRows))
	for _, row := range mappedRows {
		// os valores vêm como string; convertemos via json → struct
		b, _ := json.Marshal(row)
		var tx TransactionImportItem
		if err := json.Unmarshal(b, &tx); err != nil {
			return nil, fmt.Errorf("row to struct error: %w", err)
		}
		txs = append(txs, tx)
	}
	return txs, nil
}

// -----------------------------------------------------
// Helpers
// -----------------------------------------------------

func ParseCSV(r io.Reader) ([]map[string]any, error) {
	cr := csv.NewReader(r)
	cr.TrimLeadingSpace = true
	cr.LazyQuotes = true // Enable LazyQuotes to handle improperly quoted fields
	headers, err := cr.Read()
	if err != nil {
		return nil, fmt.Errorf("csv header: %w", err)
	}

	// Clean headers by removing quotes
	for i, h := range headers {
		// Remove all quote characters, not just at the boundaries
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

		// Pula linhas vazias
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
	// excelize precisa de io.ReadSeeker, então copiamos para buffer
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
	// header
	if !rowsIter.Next() {
		return nil, fmt.Errorf("xlsx empty sheet")
	}
	headers, _ := rowsIter.Columns()
	var rows []map[string]any
	for rowsIter.Next() {
		cols, _ := rowsIter.Columns()

		// Pula linhas vazias
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

// isEmptyRow verifica se todas as colunas de uma linha estão vazias
func isEmptyRow(cols []string) bool {
	if len(cols) == 0 {
		return true
	}

	for _, col := range cols {
		// Se pelo menos uma coluna tiver conteúdo, a linha não está vazia
		if strings.TrimSpace(col) != "" {
			return false
		}
	}

	// Se todas as colunas estiverem vazias, a linha está vazia
	return true
}

func ApplyMapping(rows []map[string]any, defs []ColumnDef) []map[string]any {
	if len(defs) == 0 {
		return rows // nada para mapear
	}

	// Pre-process header map for case-insensitive lookup
	headerMap := make(map[string]string)
	if len(rows) > 0 {
		for k := range rows[0] {
			normalizedKey := normalize(k)
			headerMap[normalizedKey] = k
		}
	}

	// Define known numeric fields that need conversion
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
		// copia tudo primeiro
		for k, v := range row {
			mappedToAppend[k] = v
		}
		// aplica mapping
		for _, col := range defs {
			normalizedKeyToMap := normalize(col.KeyToMap)

			if col.Key == "" || col.KeyToMap == "" {
				continue
			}

			if col.IsCustomField {
				// customFields é []any de objetos {id, value}
				id := col.Key
				value := row[col.KeyToMap]

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
				tags := row[col.KeyToMap].(string)
				tagsSlice, ok := mappedToAppend["tags"].([]map[string]any)
				if !ok {
					tagsSlice = []map[string]any{}
				}

				tagSplited := strings.Split(tags, ",")
				for _, tag := range tagSplited {
					splitSubTagAndTag := strings.Split(tag, "-")
					if len(splitSubTagAndTag) == 2 {
						tagsSlice = append(tagsSlice, map[string]any{
							"tag":    splitSubTagAndTag[0],
							"subTag": splitSubTagAndTag[1],
						})
					} else {
						// error here
						fmt.Println("error here")
						continue
					}
				}

				mappedToAppend["tags"] = tagsSlice
				continue
			}

			// Check if key is a nested path (using dot notation)
			if strings.Contains(col.Key, ".") {
				parts := strings.SplitN(col.Key, ".", 2)
				parentKey := parts[0]
				childKey := parts[1]

				// Find the source value
				var foundValue any
				var found bool

				// Tenta encontrar a chave original usando o mapa normalizado
				if originalKey, ok := headerMap[normalizedKeyToMap]; ok {
					if val, ok := row[originalKey]; ok {
						foundValue = val
						found = true
					}
				}

				// Busca caso-insensitiva como fallback
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
					// Convert string to float for numeric fields
					if strVal, isStr := foundValue.(string); isStr && numericFields[col.Key] {
						cleanVal := strings.ReplaceAll(strVal, ",", ".") // Convert decimal comma to dot

						if cleanVal == "" {
							cleanVal = "0"
						}

						// Parse string to float64
						floatVal, err := strconv.ParseFloat(cleanVal, 64)
						if err != nil {
							continue
						}

						foundValue = floatVal
					}

					// Ensure parent object exists
					parentObj, ok := mappedToAppend[parentKey].(map[string]any)
					if !ok {
						// Create parent object if it doesn't exist
						parentObj = make(map[string]any)
						mappedToAppend[parentKey] = parentObj
					}

					// Set the value in the nested object
					parentObj[childKey] = foundValue
				}
				continue
			}

			// Process regular non-nested fields (existing code)
			// Tenta encontrar a chave original usando o mapa normalizado
			if originalKey, ok := headerMap[normalizedKeyToMap]; ok {
				if val, ok := row[originalKey]; ok {
					mappedToAppend[col.Key] = val
					if col.Key != originalKey {
						delete(mappedToAppend, originalKey)
					}
					continue
				}
			}

			// Busca caso-insensitiva como fallback
			found := false
			var foundKey string
			var foundValue any

			for k, v := range row {
				// Antes era apenas case insensitive, agora removemos todos os caracteres não alfanuméricos
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

// Função auxiliar para normalizar strings para comparação
func normalize(s string) string {
	// Converte para minúsculas
	s = strings.ToLower(s)

	// Remove espaços extras
	s = strings.TrimSpace(s)

	// Remove caracteres invisíveis e não imprimíveis (como BOM, zero-width spaces, etc)
	var result []rune
	for _, r := range s {
		if r > 32 && r < 127 { // Mantém apenas ASCII imprimível
			result = append(result, r)
		}
	}

	return string(result)
}

func (c *ImportTransactionController) ParseAllDatesAndTypes(transactions []TransactionImportItem) ([]TransactionImportItem, error) {
	for i := range transactions {
		transactions[i].DueDate = strings.ReplaceAll(transactions[i].DueDate, "-", "/")
		transactions[i].RegistrationDate = strings.ReplaceAll(transactions[i].RegistrationDate, "-", "/")
		if transactions[i].ConfirmationDate != nil {
			*transactions[i].ConfirmationDate = strings.ReplaceAll(*transactions[i].ConfirmationDate, "-", "/")
		}

		// Parse dueDate
		if transactions[i].DueDate != "" {
			// Normalizar formato da data com dígitos únicos para o formato com dois dígitos
			parts := strings.Split(transactions[i].DueDate, "/")
			if len(parts) == 3 {
				// Adicionar zero à esquerda se necessário
				for j := 0; j < 2; j++ { // Apenas para dia e mês
					if len(parts[j]) == 1 {
						parts[j] = "0" + parts[j]
					}
				}
				transactions[i].DueDate = strings.Join(parts, "/")
			}

			// Tenta primeiro com o formato de 4 dígitos para o ano
			t, err := time.Parse("02/01/2006", transactions[i].DueDate)
			if err != nil {
				// Se falhar, tenta com o formato de 2 dígitos para o ano
				t, err = time.Parse("02/01/06", transactions[i].DueDate)
				if err != nil {
					return nil, fmt.Errorf("invalid dueDate format for transaction %d: %w", i+1, err)
				}
			}
			transactions[i].DueDate = t.UTC().Format("2006-01-02T15:04:05Z")
		}

		// Parse registrationDate
		if transactions[i].RegistrationDate != "" {
			// Normalizar formato da data com dígitos únicos para o formato com dois dígitos
			parts := strings.Split(transactions[i].RegistrationDate, "/")
			if len(parts) == 3 {
				// Adicionar zero à esquerda se necessário
				for j := 0; j < 2; j++ { // Apenas para dia e mês
					if len(parts[j]) == 1 {
						parts[j] = "0" + parts[j]
					}
				}
				transactions[i].RegistrationDate = strings.Join(parts, "/")
			}

			// Tenta primeiro com o formato de 4 dígitos para o ano
			t, err := time.Parse("02/01/2006", transactions[i].RegistrationDate)
			if err != nil {
				// Se falhar, tenta com o formato de 2 dígitos para o ano
				t, err = time.Parse("02/01/06", transactions[i].RegistrationDate)
				if err != nil {
					return nil, fmt.Errorf("invalid registrationDate format for transaction %d: %w", i+1, err)
				}
			}
			transactions[i].RegistrationDate = t.UTC().Format("2006-01-02T15:04:05Z")
		}

		// Parse confirmationDate if present
		if transactions[i].ConfirmationDate != nil && *transactions[i].ConfirmationDate != "" {
			// Normalizar formato da data com dígitos únicos para o formato com dois dígitos
			parts := strings.Split(*transactions[i].ConfirmationDate, "/")
			if len(parts) == 3 {
				// Adicionar zero à esquerda se necessário
				for j := 0; j < 2; j++ { // Apenas para dia e mês
					if len(parts[j]) == 1 {
						parts[j] = "0" + parts[j]
					}
				}
				*transactions[i].ConfirmationDate = strings.Join(parts, "/")
			}

			// Tenta primeiro com o formato de 4 dígitos para o ano
			t, err := time.Parse("02/01/2006", *transactions[i].ConfirmationDate)
			if err != nil {
				// Se falhar, tenta com o formato de 2 dígitos para o ano
				t, err = time.Parse("02/01/06", *transactions[i].ConfirmationDate)
				if err != nil {
					return nil, fmt.Errorf("invalid confirmationDate format for transaction %d: %w", i+1, err)
				}
			}
			formattedDate := t.UTC().Format("2006-01-02T15:04:05Z")
			transactions[i].ConfirmationDate = &formattedDate
			transactions[i].IsConfirmed = true
		}

		// Verifica se ConfirmationDate é nil antes de acessar
		if transactions[i].ConfirmationDate != nil && *transactions[i].ConfirmationDate == "" {
			transactions[i].ConfirmationDate = nil
			transactions[i].IsConfirmed = false
		}

		transactions[i].Frequency = "DO_NOT_REPEAT"

		if strings.ToLower(transactions[i].Type) == "receita" || strings.ToLower(transactions[i].Type) == "recurring" {
			transactions[i].Type = "RECIPE"
		}

		if strings.ToLower(transactions[i].Type) == "despesa" || strings.ToLower(transactions[i].Type) == "expense" {
			transactions[i].Type = "EXPENSE"
		}
	}
	return transactions, nil
}

// translateErrorMessage traduz mensagens de erro para português
func (c *ImportTransactionController) translateErrorMessage(errorMsg string) string {
	// Mapeamento de mensagens de erro em inglês para português
	errorTranslations := map[string]string{
		"member not found with email":                  "membro não encontrado com o email",
		"bank not found":                               "banco não encontrado",
		"error creating account":                       "erro ao criar conta",
		"error creating category":                      "erro ao criar categoria",
		"error updating category with new subcategory": "erro ao atualizar categoria com nova subcategoria",
		"custom field not found":                       "campo personalizado não encontrado",
		"custom field type mismatch":                   "tipo de campo personalizado incompatível",
		"invalid custom field ID":                      "ID de campo personalizado inválido",
		"category is not a tag":                        "categoria não é uma tag",
		"error creating tag":                           "erro ao criar tag",
		"error updating tag with new subtag":           "erro ao atualizar tag com nova subtag",
		"panic recovered":                              "erro inesperado",
		"invalid user ID format":                       "formato de ID de usuário inválido",
		"invalid workspace ID format":                  "formato de ID de espaço de trabalho inválido",
	}

	// Procura por fragmentos de mensagens de erro conhecidas e substitui
	for engMsg, ptMsg := range errorTranslations {
		if strings.Contains(errorMsg, engMsg) {
			// Substitui a parte em inglês pelo equivalente em português
			return strings.Replace(errorMsg, engMsg, ptMsg, 1)
		}
	}

	// Se não encontrar uma tradução específica, retorna a mensagem original
	return errorMsg
}

// translateValidationError traduz erros de validação para português
func (c *ImportTransactionController) translateValidationError(err validator.FieldError) string {
	fieldName := err.Field()
	tag := err.Tag()
	param := err.Param()

	// Remover prefixo "Transactions[n]." se existir
	if strings.Contains(fieldName, "Transactions") && strings.Contains(fieldName, ".") {
		parts := strings.Split(fieldName, ".")
		if len(parts) > 1 {
			fieldName = parts[len(parts)-1]
		}
	}

	// Mapeamento de nomes de campos para português
	fieldTranslations := map[string]string{
		"Name":        "Nome",
		"Description": "Descrição",
		"Invoice":     "Fatura",
		"Type":        "Tipo",
		"Supplier":    "Fornecedor",
		"AssignedTo":  "Responsável",
		"Balance":     "Balanço",
		"Value":       "Valor",
		"Discount":    "Desconto",
		"Interest":    "Juros",
		"DueDate":     "Data de vencimento",
		"Account":     "Conta",
		"Category":    "Categoria",
		"SubCategory": "Subcategoria",
	}

	// Tradução do nome do campo
	fieldTranslated := fieldName
	if translated, ok := fieldTranslations[fieldName]; ok {
		fieldTranslated = translated
	}

	// Mapeamento de mensagens de erro de validação
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
		return fieldTranslated + " deve estar no formato de data válido"
	default:
		return "Erro de validação no campo " + fieldTranslated + ": " + tag
	}
}
