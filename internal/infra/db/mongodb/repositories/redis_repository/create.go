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

func SaveCSVToRedis(redisURL string, key string, data [][]string, expiration time.Duration) error {
	redisClient := helpers.RedisHelper(redisURL)
	ctx, cancel := context.WithTimeout(context.Background(), helpers.RedisTimeout)
	defer cancel()

	buf := new(bytes.Buffer)
	writer := csv.NewWriter(buf)

	for _, record := range data {
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("erro ao escrever registro CSV: %w", err)
		}
	}
	writer.Flush()

	if err := writer.Error(); err != nil {
		return fmt.Errorf("erro ao gerar CSV: %w", err)
	}

	err := redisClient.Set(ctx, key, buf.String(), expiration).Err()
	if err != nil {
		return fmt.Errorf("erro ao salvar CSV no Redis: %w", err)
	}

	return nil
}

func SaveExcelToRedis(redisURL string, key string, excelData *excelize.File, expiration time.Duration) error {
	redisClient := helpers.RedisHelper(redisURL)
	ctx, cancel := context.WithTimeout(context.Background(), helpers.RedisTimeout)
	defer cancel()

	buf := new(bytes.Buffer)
	if err := excelData.Write(buf); err != nil {
		return fmt.Errorf("erro ao serializar Excel: %w", err)
	}

	encodedData := base64.StdEncoding.EncodeToString(buf.Bytes())

	err := redisClient.Set(ctx, key, encodedData, expiration).Err()
	if err != nil {
		return fmt.Errorf("erro ao salvar Excel no Redis: %w", err)
	}

	return nil
}
