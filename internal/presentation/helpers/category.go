package helpers

import (
	"github.com/anuntech/finance-backend/internal/domain/models"
)

// CheckDuplicateSubCategories verifica se há subcategorias com nomes duplicados
// Retorna true se houver duplicatas, false se todos os nomes forem únicos
func CheckDuplicateSubCategories(subCategories []models.SubCategoryCategory) bool {
	nameSet := make(map[string]bool)

	for _, subCat := range subCategories {
		if nameSet[subCat.Name] {
			return true
		}
		nameSet[subCat.Name] = true
	}

	return false
}
