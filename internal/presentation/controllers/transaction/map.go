package transaction

import (
	"encoding/json"
	"net/http"

	"github.com/anuntech/finance-backend/internal/presentation/helpers"
	presentationProtocols "github.com/anuntech/finance-backend/internal/presentation/protocols"
)

// MapRequest representa o corpo da requisição para o mapeamento
type MapRequest struct {
	Columns []struct {
		Key           string `json:"key"`           // vai ser o novo nome da key
		KeyToMap      string `json:"keyToMap"`      // vai ser a chave que vair vir
		IsCustomField bool   `json:"isCustomField"` // true ou false
	} `json:"columns"`
	Data []map[string]any `json:"data"` // all transactions
}

// MapTransactionController controlador para mapear transações
type MapTransactionController struct{}

// NewMapTransactionController cria uma nova instância do controlador
func NewMapTransactionController() *MapTransactionController {
	return &MapTransactionController{}
}

// Handle processa a requisição HTTP
func (c *MapTransactionController) Handle(r presentationProtocols.HttpRequest) *presentationProtocols.HttpResponse {
	var body MapRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "invalid body request: " + err.Error(),
		}, http.StatusBadRequest)
	}

	// Verifica se há dados para processar
	if len(body.Data) == 0 {
		return helpers.CreateResponse([]map[string]any{}, http.StatusOK)
	}

	// Processa o mapeamento
	mappedData := make([]map[string]any, len(body.Data))

	for i, transaction := range body.Data {
		// Inicializa com uma cópia da transação original para manter todos os campos
		mappedTransaction := make(map[string]any)
		for k, v := range transaction {
			mappedTransaction[k] = v
		}

		for _, column := range body.Columns {
			if column.IsCustomField {
				// Processa campos customizados
				if customFields, ok := transaction["customFields"].([]any); ok {
					for _, cf := range customFields {
						if customField, ok := cf.(map[string]any); ok {
							// Assume que o customField tem uma propriedade 'id' que corresponde ao keyToMap
							if customField["id"] == column.KeyToMap {
								customField["id"] = column.Key
								break
							}
						}
					}
					mappedTransaction["customFields"] = customFields
				}
			} else {
				// Mapeamento direto de propriedades regulares
				if value, exists := transaction[column.KeyToMap]; exists {
					mappedTransaction[column.Key] = value
					// Remover o campo original apenas se o novo nome for diferente
					if column.Key != column.KeyToMap {
						delete(mappedTransaction, column.KeyToMap)
					}
				}
			}
		}

		mappedData[i] = mappedTransaction
	}

	return helpers.CreateResponse(mappedData, http.StatusOK)
}
