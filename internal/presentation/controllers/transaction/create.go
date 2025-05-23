package transaction

import (
	"encoding/json"
	"errors"
	"net/http"
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
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CreateTransactionController struct {
	Validate                      *validator.Validate
	Translator                    ut.Translator
	CreateTransactionRepository   usecase.CreateTransactionRepository
	FindMemberByIdRepository      *member_repository.FindMemberByIdRepository
	FindAccountByIdRepository     usecase.FindAccountByIdRepository
	FindCategoryByIdRepository    usecase.FindCategoryByIdRepository
	FindCustomFieldByIdRepository usecase.FindCustomFieldByIdRepository
}

func NewCreateTransactionController(findMemberByIdRepository *member_repository.FindMemberByIdRepository, createTransactionRepository *transaction_repository.CreateTransactionRepository, findAccountByIdRepository usecase.FindAccountByIdRepository, findCategoryByIdRepository usecase.FindCategoryByIdRepository, findCustomFieldByIdRepository usecase.FindCustomFieldByIdRepository) *CreateTransactionController {
	validate := validator.New(validator.WithRequiredStructEnabled())

	return &CreateTransactionController{
		Validate:                      validate,
		FindMemberByIdRepository:      findMemberByIdRepository,
		CreateTransactionRepository:   createTransactionRepository,
		FindAccountByIdRepository:     findAccountByIdRepository,
		FindCategoryByIdRepository:    findCategoryByIdRepository,
		FindCustomFieldByIdRepository: findCustomFieldByIdRepository,
	}
}

type TransactionBody struct {
	Name        string  `json:"name" validate:"required,min=2,max=30"`
	MainId      *string `json:"mainId" validate:"required_with=MainCount,omitempty,mongodb"`
	MainCount   *int    `json:"mainCount" validate:"required_with=MainId,omitempty,min=1,max=367"`
	Description string  `json:"description" validate:"omitempty,max=255"`
	Invoice     string  `json:"invoice" validate:"omitempty,min=2,max=50"`
	Type        string  `json:"type" validate:"required,oneof=EXPENSE RECIPE"`
	Supplier    string  `json:"supplier" validate:"omitempty,min=3,max=30"`
	AssignedTo  string  `json:"assignedTo" validate:"required,min=3,max=30,mongodb"`
	Balance     struct {
		Value              float64 `json:"value" validate:"required,min=0.01"`
		Parts              float64 `json:"parts" validate:"omitempty,min=0.01"`
		Labor              float64 `json:"labor" validate:"omitempty,min=0.01"`
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
		CustomDay          int        `json:"customDay" validate:"required_if=Interval CUSTOM"`
	} `json:"repeatSettings" validate:"excluded_if=Frequency DO_NOT_REPEAT,excluded_if=Frequency RECURRING,required_if=Frequency REPEAT,omitempty"`
	DueDate       string  `json:"dueDate" validate:"required,datetime=2006-01-02T15:04:05Z"`
	IsConfirmed   bool    `json:"isConfirmed"`
	CategoryId    *string `json:"categoryId" validate:"required_with=SubCategoryId,omitempty,mongodb"`
	SubCategoryId *string `json:"subCategoryId" validate:"omitempty,mongodb"`
	Tags          []struct {
		TagId    string `json:"tagId" validate:"omitempty,mongodb"`
		SubTagId string `json:"subTagId" validate:"excluded_if=TagId '',omitempty,mongodb"`
	} `json:"tags" validate:"omitempty"`
	CustomFields []struct {
		CustomFieldId string `json:"id" validate:"required,mongodb"`
		Value         string `json:"value" validate:"required,max=100"`
	} `json:"customFields"`
	AccountId        *string `json:"accountId" validate:"required,mongodb"`
	RegistrationDate string  `json:"registrationDate" validate:"required,datetime=2006-01-02T15:04:05Z"`
	ConfirmationDate *string `json:"confirmationDate" validate:"excluded_if=IsConfirmed false,required_if=IsConfirmed true,omitempty,datetime=2006-01-02T15:04:05Z,excluded_with=CreditCardId"`
	CreditCardId     *string `json:"creditCardId" validate:"omitempty,mongodb"`
}

func (c *CreateTransactionController) Handle(r presentationProtocols.HttpRequest) *presentationProtocols.HttpResponse {
	var body TransactionBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "formato da solicitação inválido",
		}, http.StatusBadRequest)
	}

	if err := c.Validate.Struct(body); err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: helpers.GetErrorMessages(c.Validate, err),
		}, http.StatusBadRequest)
	}

	transaction, err := createTransaction(&body)
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "erro ao criar a transação: " + err.Error(),
		}, http.StatusInternalServerError)
	}

	userObjectID, err := primitive.ObjectIDFromHex(r.Header.Get("userId"))
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "formato do ID do usuário inválido",
		}, http.StatusBadRequest)
	}
	transaction.CreatedBy = userObjectID

	assignedTo, err := primitive.ObjectIDFromHex(body.AssignedTo)
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "formato do ID do responsável inválido",
		}, http.StatusBadRequest)
	}

	workspaceId, err := primitive.ObjectIDFromHex(r.Header.Get("workspaceId"))
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "formato do ID da área de trabalho inválido",
		}, http.StatusBadRequest)
	}
	transaction.WorkspaceId = workspaceId

	errChan := make(chan *presentationProtocols.HttpResponse, 4)
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer utils.RecoveryWithCallback(&wg, func(r interface{}) {
			errChan <- helpers.CreateResponse(&presentationProtocols.ErrorResponse{
				Error: "erro na validação do responsável: ocorreu um erro inesperado",
			}, http.StatusInternalServerError)
		})
		defer wg.Done()

		if err := c.validateAssignedMember(workspaceId, assignedTo); err != nil {
			errChan <- err
			return
		}
		transaction.AssignedTo = assignedTo
	}()

	wg.Add(1)
	go func() {
		defer utils.RecoveryWithCallback(&wg, func(r interface{}) {
			errChan <- helpers.CreateResponse(&presentationProtocols.ErrorResponse{
				Error: "erro na validação da conta: ocorreu um erro inesperado",
			}, http.StatusInternalServerError)
		})
		defer wg.Done()
		if transaction.AccountId == nil {
			return
		}
		if err := c.validateAccount(workspaceId, *transaction.AccountId); err != nil {
			errChan <- err
		}
	}()

	wg.Add(1)
	go func() {
		defer utils.RecoveryWithCallback(&wg, func(r interface{}) {
			errChan <- helpers.CreateResponse(&presentationProtocols.ErrorResponse{
				Error: "erro na validação da categoria: ocorreu um erro inesperado",
			}, http.StatusInternalServerError)
		})
		defer wg.Done()

		// Se não tiver categoria, não valida
		if transaction.CategoryId == nil {
			return
		}

		// Garante que categoryId não é nil antes de desreferenciar
		categoryId := *transaction.CategoryId

		// Se não tiver subcategoria, passa o ID nulo
		var subCategoryId primitive.ObjectID
		if transaction.SubCategoryId != nil {
			subCategoryId = *transaction.SubCategoryId
		}

		// Faz a validação com os valores extraídos
		if err := c.validateCategory(workspaceId, categoryId, transaction.Type, subCategoryId); err != nil {
			errChan <- err
		}
	}()

	wg.Add(1)
	go func() {
		defer utils.RecoveryWithCallback(&wg, func(r interface{}) {
			errChan <- helpers.CreateResponse(&presentationProtocols.ErrorResponse{
				Error: "erro na validação dos campos personalizados: ocorreu um erro inesperado",
			}, http.StatusInternalServerError)
		})
		defer wg.Done()

		seenCustomFields := make(map[string]bool)

		for _, customField := range transaction.CustomFields {
			if customField.CustomFieldId.IsZero() {
				errChan <- helpers.CreateResponse(&presentationProtocols.ErrorResponse{
					Error: "ID do campo personalizado inválido",
				}, http.StatusBadRequest)
				return
			}

			compositeKey := customField.CustomFieldId.Hex()

			if seenCustomFields[compositeKey] {
				errChan <- helpers.CreateResponse(&presentationProtocols.ErrorResponse{
					Error: "campo personalizado duplicado detectado: " + compositeKey,
				}, http.StatusBadRequest)
				return
			}
			seenCustomFields[compositeKey] = true

			customFieldParsed, err := c.FindCustomFieldByIdRepository.Find(customField.CustomFieldId, workspaceId)
			if err != nil {
				errChan <- helpers.CreateResponse(&presentationProtocols.ErrorResponse{
					Error: "erro ao buscar campo personalizado",
				}, http.StatusInternalServerError)
				return
			}

			if customFieldParsed == nil {
				errChan <- helpers.CreateResponse(&presentationProtocols.ErrorResponse{
					Error: "campo personalizado não encontrado",
				}, http.StatusNotFound)
				return
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer utils.RecoveryWithCallback(&wg, func(r interface{}) {
			errChan <- helpers.CreateResponse(&presentationProtocols.ErrorResponse{
				Error: "erro na validação das tags: ocorreu um erro inesperado",
			}, http.StatusInternalServerError)
		})
		defer wg.Done()
		seenTags := make(map[string]bool)

		for _, tag := range transaction.Tags {
			if tag.SubTagId == primitive.NilObjectID {
				errChan <- helpers.CreateResponse(&presentationProtocols.ErrorResponse{
					Error: "subtag não encontrada",
				}, http.StatusNotFound)
				return
			}

			compositeKey := tag.TagId.Hex() + "|" + tag.SubTagId.Hex()

			if seenTags[compositeKey] {
				errChan <- helpers.CreateResponse(&presentationProtocols.ErrorResponse{
					Error: "tag duplicada detectada: " + compositeKey,
				}, http.StatusBadRequest)
				return
			}
			seenTags[compositeKey] = true

			if err := c.validateTag(workspaceId, tag.TagId, tag.SubTagId); err != nil {
				errChan <- err
				return
			}
		}
	}()

	wg.Wait()
	close(errChan)

	if len(errChan) > 0 {
		return <-errChan
	}

	transaction, err = c.CreateTransactionRepository.Create(transaction)
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "erro ao criar a transação",
		}, http.StatusInternalServerError)
	}

	recipeTx := *transaction
	recipeTx.Type = "RECIPE"
	recipeNetBalance := infraHelpers.CalculateOneTransactionBalance(&recipeTx)
	transaction.Balance.NetBalance = recipeNetBalance

	return helpers.CreateResponse(transaction, http.StatusCreated)
}

func createTransaction(body *TransactionBody) (*models.Transaction, error) {
	if body.Frequency == "REPEAT" && int(body.RepeatSettings.InitialInstallment) >= body.RepeatSettings.Count {
		return nil, errors.New("initialInstallment must be less than count")
	}

	convertID := func(id string) (primitive.ObjectID, error) {
		return primitive.ObjectIDFromHex(id)
	}

	parseDate := func(date string) (time.Time, error) {
		location := time.UTC
		return time.ParseInLocation("2006-01-02T15:04:05Z", date, location)
	}

	var categoryId *primitive.ObjectID
	if body.CategoryId != nil {
		categoryIdParsed, err := convertID(*body.CategoryId)
		if err != nil {
			return nil, err
		}

		categoryId = &categoryIdParsed
	}

	var subCategoryId *primitive.ObjectID
	if body.SubCategoryId != nil {
		subCategoryIdParsed, err := convertID(*body.SubCategoryId)
		if err != nil {
			return nil, err
		}

		subCategoryId = &subCategoryIdParsed
	}

	var tags = []models.TransactionTags{}

	for _, tag := range body.Tags {
		var tagId primitive.ObjectID
		if tag.TagId != "" {
			tagIdParsed, err := convertID(tag.TagId)
			if err != nil {
				return nil, err
			}

			tagId = tagIdParsed
		}

		var subTagId primitive.ObjectID
		if tag.SubTagId != "" {
			subTagIdParsed, err := convertID(tag.SubTagId)
			if err != nil {
				return nil, err
			}

			subTagId = subTagIdParsed
		}

		tags = append(tags, models.TransactionTags{
			TagId:    tagId,
			SubTagId: subTagId,
		})
	}

	assignedTo, err := convertID(body.AssignedTo)
	if err != nil {
		return nil, err
	}

	var accountId *primitive.ObjectID
	if body.AccountId != nil {
		accountIdParsed, err := convertID(*body.AccountId)
		if err != nil {
			return nil, err
		}

		accountId = &accountIdParsed
	}

	registrationDate, err := parseDate(body.RegistrationDate)
	if err != nil {
		return nil, err
	}

	var confirmationDate *time.Time
	if body.IsConfirmed {
		confirmationDateParsed, err := parseDate(*body.ConfirmationDate)
		if err != nil {
			return nil, err
		}

		confirmationDate = &confirmationDateParsed
	} else {
		confirmationDate = nil
	}

	dueDate, err := parseDate(body.DueDate)
	if err != nil {
		return nil, err
	}

	var customFields = []models.TransactionCustomField{}

	for _, customField := range body.CustomFields {
		customFieldIdParsed, err := convertID(customField.CustomFieldId)
		if err != nil {
			return nil, err
		}

		customFields = append(customFields, models.TransactionCustomField{
			CustomFieldId: customFieldIdParsed,
			Value:         customField.Value,
		})
	}

	return &models.Transaction{
		Name:        body.Name,
		Description: body.Description,
		Invoice:     body.Invoice,
		Type:        body.Type,
		Supplier:    body.Supplier,
		AssignedTo:  assignedTo,
		Balance: models.TransactionBalance{
			Value:              body.Balance.Value,
			Discount:           body.Balance.Discount,
			Interest:           body.Balance.Interest,
			DiscountPercentage: body.Balance.DiscountPercentage,
			InterestPercentage: body.Balance.InterestPercentage,
		},
		Frequency: body.Frequency,
		RepeatSettings: &models.TransactionRepeatSettings{
			InitialInstallment: body.RepeatSettings.InitialInstallment,
			Count:              body.RepeatSettings.Count,
			Interval:           body.RepeatSettings.Interval,
			CustomDay:          body.RepeatSettings.CustomDay,
		},
		CustomFields:     customFields,
		IsConfirmed:      body.IsConfirmed,
		CategoryId:       categoryId,
		SubCategoryId:    subCategoryId,
		Tags:             tags,
		AccountId:        accountId,
		RegistrationDate: registrationDate,
		ConfirmationDate: confirmationDate,
		DueDate:          dueDate,
	}, nil
}

func (c *CreateTransactionController) validateAssignedMember(workspaceId primitive.ObjectID, assignedTo primitive.ObjectID) *presentationProtocols.HttpResponse {
	member, err := c.FindMemberByIdRepository.Find(workspaceId, assignedTo)
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "erro ao buscar membro",
		}, http.StatusInternalServerError)
	}

	if member == nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "o responsável não é um membro da área de trabalho",
		}, http.StatusNotFound)
	}

	return nil
}

func (c *CreateTransactionController) validateAccount(workspaceId primitive.ObjectID, accountId primitive.ObjectID) *presentationProtocols.HttpResponse {
	account, err := c.FindAccountByIdRepository.Find(accountId, workspaceId)
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "erro ao buscar conta",
		}, http.StatusInternalServerError)
	}

	if account == nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "conta não encontrada",
		}, http.StatusNotFound)
	}

	return nil
}

func (c *CreateTransactionController) validateCategory(workspaceId primitive.ObjectID, categoryId primitive.ObjectID, transactionType string, subCategoryId primitive.ObjectID) *presentationProtocols.HttpResponse {
	category, err := c.FindCategoryByIdRepository.Find(categoryId, workspaceId)
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "erro ao buscar categoria",
		}, http.StatusInternalServerError)
	}

	if category == nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "categoria não encontrada",
		}, http.StatusNotFound)
	}

	if !strings.EqualFold(category.Type, transactionType) {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "o tipo da categoria não corresponde ao tipo da transação",
		}, http.StatusBadRequest)
	}

	if subCategoryId == primitive.NilObjectID {
		return nil
	}

	for _, subCategory := range category.SubCategories {
		if subCategory.Id == subCategoryId {
			return nil
		}
	}

	return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
		Error: "subcategoria não encontrada",
	}, http.StatusNotFound)
}

func (c *CreateTransactionController) validateTag(workspaceId primitive.ObjectID, categoryId primitive.ObjectID, subCategoryId primitive.ObjectID) *presentationProtocols.HttpResponse {
	category, err := c.FindCategoryByIdRepository.Find(categoryId, workspaceId)
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "erro ao buscar tag",
		}, http.StatusInternalServerError)
	}

	if category == nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "tag não encontrada",
		}, http.StatusNotFound)
	}

	if !strings.EqualFold(category.Type, "TAG") {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "o tipo da tag não corresponde ao tipo da transação",
		}, http.StatusBadRequest)
	}

	for _, subCategory := range category.SubCategories {
		if subCategory.Id == subCategoryId {
			return nil
		}
	}

	if subCategoryId == primitive.NilObjectID {
		return nil
	}

	return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
		Error: "subtag não encontrada",
	}, http.StatusNotFound)
}
