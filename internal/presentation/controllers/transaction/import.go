package transaction

import (
	"encoding/json"
	"errors"
	"fmt"
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
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Interfaces auxiliares para busca por nome
type FindMemberByEmailRepository interface {
	FindByEmailAndWorkspaceId(email string, workspaceId primitive.ObjectID) (*models.Member, error)
	// Mantém o método antigo para compatibilidade
	FindByNameAndWorkspaceId(name string, workspaceId primitive.ObjectID) (*models.Member, error)
}

type FindCustomFieldByNameRepository interface {
	FindByNameAndWorkspaceId(name string, workspaceId primitive.ObjectID) (*models.CustomField, error)
}

type ImportTransactionController struct {
	Validate                      *validator.Validate
	Translator                    ut.Translator
	CreateTransactionRepository   usecase.CreateTransactionRepository
	FindMemberByIdRepository      *member_repository.FindMemberByIdRepository
	FindAccountByIdRepository     usecase.FindAccountByIdRepository
	FindCategoryByIdRepository    usecase.FindCategoryByIdRepository
	FindCustomFieldByIdRepository usecase.FindCustomFieldByIdRepository
	// Repositórios para busca por nome
	FindAccountByNameRepository         usecase.FindAccountByNameAndWorkspaceIdRepository
	FindCategoryByNameAndTypeRepository usecase.FindCategoryByNameAndTypeRepository
	FindMemberByEmailRepository         FindMemberByEmailRepository
	FindCustomFieldByNameRepository     FindCustomFieldByNameRepository
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
	}
}

type TransactionImportItem struct {
	Name        string `json:"name" validate:"required,min=2,max=30"`
	Description string `json:"description" validate:"omitempty,max=255"`
	Invoice     string `json:"invoice" validate:"omitempty,min=2,max=50"`
	Type        string `json:"type" validate:"required,oneof=EXPENSE RECIPE"`
	Supplier    string `json:"supplier" validate:"omitempty,min=3,max=30"`
	AssignedTo  string `json:"assignedTo" validate:"required,email"` // Email do membro
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
		Interval           string     `json:"interval" validate:"oneof=DAILY WEEKLY MONTHLY QUARTERLY YEARLY"`
	} `json:"repeatSettings" validate:"excluded_if=Frequency DO_NOT_REPEAT,excluded_if=Frequency RECURRING,required_if=Frequency REPEAT,omitempty"`
	DueDate     string  `json:"dueDate" validate:"required,datetime=2006-01-02T15:04:05Z"`
	IsConfirmed bool    `json:"isConfirmed"`
	Category    *string `json:"category" validate:"omitempty"`    // Nome da categoria
	SubCategory *string `json:"subCategory" validate:"omitempty"` // Nome da subcategoria
	Tags        []struct {
		Tag    string `json:"tag" validate:"omitempty"`    // Nome da tag
		SubTag string `json:"subTag" validate:"omitempty"` // Nome da subtag
	} `json:"tags" validate:"omitempty"`
	CustomFields []struct {
		CustomField string `json:"customField" validate:"required"` // Nome do campo personalizado
		Value       string `json:"value" validate:"required,max=100"`
	} `json:"customFields" validate:"omitempty"`
	Account          string  `json:"account" validate:"required"` // Nome da conta
	RegistrationDate string  `json:"registrationDate" validate:"required,datetime=2006-01-02T15:04:05Z"`
	ConfirmationDate *string `json:"confirmationDate" validate:"excluded_if=IsConfirmed false,required_if=IsConfirmed true,omitempty,datetime=2006-01-02T15:04:05Z"`
}

type ImportTransactionBody struct {
	Transactions []TransactionImportItem `json:"transactions" validate:"required,dive"`
}

func (c *ImportTransactionController) Handle(r presentationProtocols.HttpRequest) *presentationProtocols.HttpResponse {
	var body ImportTransactionBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "invalid body request",
		}, http.StatusBadRequest)
	}

	if err := c.Validate.Struct(body); err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: helpers.GetErrorMessages(c.Validate, err),
		}, http.StatusBadRequest)
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

	// Limite de transações
	if len(body.Transactions) > 100 {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "maximum of 100 transactions per import",
		}, http.StatusBadRequest)
	}

	var wg sync.WaitGroup
	importedTransactions := make([]*models.Transaction, len(body.Transactions))
	type errorInfo struct {
		index int
		err   error
	}
	errs := make(chan errorInfo, len(body.Transactions))

	// Iniciando uma goroutine para cada transação
	for i, txImport := range body.Transactions {
		wg.Add(1)
		go func(index int, tx TransactionImportItem) {
			defer wg.Done()

			// Converte a transação importada para o modelo interno
			transaction, err := c.convertImportedTransaction(&tx, workspaceId, userObjectID)
			if err != nil {
				errs <- errorInfo{index: index, err: err}
				return
			}

			// Cria a transação
			createdTx, err := c.CreateTransactionRepository.Create(transaction)
			if err != nil {
				errs <- errorInfo{index: index, err: fmt.Errorf("error creating transaction: %w", err)}
				return
			}

			// Calcula o balanço líquido
			recipeTx := *createdTx
			recipeTx.Type = "RECIPE"
			recipeNetBalance := infraHelpers.CalculateOneTransactionBalance(&recipeTx)
			createdTx.Balance.NetBalance = recipeNetBalance

			// Armazena o resultado no slice na posição correta
			importedTransactions[index] = createdTx
		}(i, txImport)
	}

	// Goroutine para monitorar erros e fechar o canal após todas as tarefas finalizarem
	go func() {
		wg.Wait()
		close(errs)
	}()

	// Verifica se ocorreu algum erro
	for e := range errs {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: fmt.Sprintf("error processing transaction #%d: %s", e.index+1, e.err.Error()),
		}, http.StatusBadRequest)
	}

	// Remove valores nil do slice de resultados (caso alguma goroutine tenha terminado com erro)
	finalTransactions := make([]*models.Transaction, 0, len(importedTransactions))
	for _, tx := range importedTransactions {
		if tx != nil {
			finalTransactions = append(finalTransactions, tx)
		}
	}

	return helpers.CreateResponse(finalTransactions, http.StatusCreated)
}

func (c *ImportTransactionController) convertImportedTransaction(txImport *TransactionImportItem, workspaceId, userID primitive.ObjectID) (*models.Transaction, error) {
	// Parse dates
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

	// Buscar membro por email
	member, err := c.FindMemberByEmailRepository.FindByEmailAndWorkspaceId(txImport.AssignedTo, workspaceId)
	if err != nil {
		return nil, err
	}
	if member == nil {
		return nil, errors.New("member not found with email: " + txImport.AssignedTo)
	}

	// Buscar conta por nome
	account, err := c.FindAccountByNameRepository.FindByNameAndWorkspaceId(txImport.Account, workspaceId)
	if err != nil {
		return nil, err
	}
	if account == nil {
		return nil, errors.New("account not found: " + txImport.Account)
	}

	// Buscar categoria e subcategoria por nome
	var categoryId *primitive.ObjectID
	var subCategoryId *primitive.ObjectID

	if txImport.Category != nil && *txImport.Category != "" {
		category, err := c.FindCategoryByNameAndTypeRepository.Find(*txImport.Category, txImport.Type, workspaceId)
		if err != nil {
			return nil, err
		}
		if category == nil {
			return nil, errors.New("category not found: " + *txImport.Category)
		}

		categoryId = &category.Id

		// Buscar subcategoria se fornecida
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
				return nil, errors.New("subcategory not found: " + *txImport.SubCategory)
			}
		}
	}

	// Processar campos personalizados
	customFields := make([]models.TransactionCustomField, 0)
	for _, cf := range txImport.CustomFields {
		customField, err := c.FindCustomFieldByNameRepository.FindByNameAndWorkspaceId(cf.CustomField, workspaceId)
		if err != nil {
			return nil, err
		}
		if customField == nil {
			return nil, errors.New("custom field not found: " + cf.CustomField)
		}

		// Assume IDs como ObjectIDs
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

	// Processar tags
	tags := make([]models.TransactionTags, 0)
	for _, tag := range txImport.Tags {
		if tag.Tag == "" {
			continue
		}

		category, err := c.FindCategoryByNameAndTypeRepository.Find(tag.Tag, "TAG", workspaceId)
		if err != nil {
			return nil, err
		}
		if category == nil {
			return nil, errors.New("tag not found: " + tag.Tag)
		}

		if !strings.EqualFold(category.Type, "TAG") {
			return nil, errors.New("category is not a tag: " + tag.Tag)
		}

		tags = append(tags, models.TransactionTags{
			TagId:    category.Id,
			SubTagId: primitive.NilObjectID, // Inicializa como nulo
		})

		// Processa subtag se fornecida
		if tag.SubTag != "" {
			found := false
			for _, subCat := range category.SubCategories {
				if strings.EqualFold(subCat.Name, tag.SubTag) {
					// Atualiza o último elemento adicionado
					tags[len(tags)-1].SubTagId = subCat.Id
					found = true
					break
				}
			}

			if !found {
				return nil, errors.New("subtag not found: " + tag.SubTag)
			}
		}
	}

	// Criar a estrutura de RepeatSettings
	var repeatSettings *models.TransactionRepeatSettings
	if txImport.Frequency == "REPEAT" {
		repeatSettings = &models.TransactionRepeatSettings{
			InitialInstallment: txImport.RepeatSettings.InitialInstallment,
			Count:              txImport.RepeatSettings.Count,
			Interval:           txImport.RepeatSettings.Interval,
		}
	}

	// Montar a transação final
	transaction := &models.Transaction{
		Id:          primitive.NewObjectID(),
		Name:        txImport.Name,
		Description: txImport.Description,
		Invoice:     txImport.Invoice,
		Type:        txImport.Type,
		Supplier:    txImport.Supplier,
		AssignedTo:  member.ID,
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
