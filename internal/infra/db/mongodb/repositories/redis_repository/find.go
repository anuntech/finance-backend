package redis_repository

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/csv"
	"fmt"
	"time"

	"github.com/xuri/excelize/v2"

	"github.com/anuntech/finance-backend/internal/infra/db/mongodb/helpers"
)

func FindByKey(redisURL string, key string) (string, error) {
	redisClient := helpers.RedisHelper(redisURL)
	ctx, cancel := context.WithTimeout(context.Background(), helpers.RedisTimeout)
	defer cancel()

	value, err := redisClient.Get(ctx, key).Result()
	if err != nil {
		return "", fmt.Errorf("erro ao buscar chave %s no Redis: %w", key, err)
	}

	return value, nil
}

func FindCSVByKey(redisURL string, key string) ([][]string, error) {
	csvString, err := FindByKey(redisURL, key)
	if err != nil {
		return nil, err
	}

	reader := csv.NewReader(bytes.NewBufferString(csvString))
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("erro ao converter string para CSV: %w", err)
	}

	return records, nil
}

func FindExcelByKey(redisURL string, key string) (*excelize.File, error) {
	encodedExcel, err := FindByKey(redisURL, key)
	if err != nil {
		return nil, err
	}

	excelBytes, err := base64.StdEncoding.DecodeString(encodedExcel)
	if err != nil {
		return nil, fmt.Errorf("erro ao decodificar Excel em base64: %w", err)
	}

	reader := bytes.NewReader(excelBytes)
	excelFile, err := excelize.OpenReader(reader)
	if err != nil {
		return nil, fmt.Errorf("erro ao abrir arquivo Excel: %w", err)
	}

	return excelFile, nil
}

func KeyExists(redisURL string, key string) (bool, error) {
	redisClient := helpers.RedisHelper(redisURL)
	ctx, cancel := context.WithTimeout(context.Background(), helpers.RedisTimeout)
	defer cancel()

	exists, err := redisClient.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("erro ao verificar existência da chave %s: %w", key, err)
	}

	return exists > 0, nil
}

func GetTTL(redisURL string, key string) (time.Duration, error) {
	redisClient := helpers.RedisHelper(redisURL)
	ctx, cancel := context.WithTimeout(context.Background(), helpers.RedisTimeout)
	defer cancel()

	ttl, err := redisClient.TTL(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("erro ao obter TTL da chave %s: %w", key, err)
	}

	return ttl, nil
}
