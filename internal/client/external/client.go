package external

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"go_sql_mid_trainer_v2/internal/domain"
)

type Client struct {
	baseURL string
	client  *http.Client
}

func New(baseURL string, client *http.Client) *Client {
	if client == nil {
		client = &http.Client{Timeout: 3 * time.Second}
	}
	return &Client{baseURL: baseURL, client: client}
}

// TODO #11.
// Реализуй HTTP-клиент внешнего risk-сервиса.
// Требования:
//   - если userID <= 0, верни domain.ErrWrongID;
//   - url = c.baseURL + "/external/risk/" + strconv.FormatInt(userID, 10);
//   - http.NewRequestWithContext;
//   - c.client.Do(req);
//   - сразу defer resp.Body.Close();
//   - если статус не 200:
//   - прочитай не больше 4KB body через io.LimitReader;
//   - верни ошибку с HTTP status и кусочком body;
//   - не используй %w, если нет исходной ошибки;
//   - json.NewDecoder(resp.Body).Decode(&profile);
//   - ошибки создания запроса, Do и Decode оборачивай через %w.
func (c *Client) GetRiskProfile(ctx context.Context, userID int64) (domain.RiskProfile, error) {
	if userID <= 0 {
		return domain.RiskProfile{}, domain.ErrWrongID
	}

	url := c.baseURL + "/external/risk/" + strconv.FormatInt(userID, 10)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return domain.RiskProfile{}, fmt.Errorf("new request err: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return domain.RiskProfile{}, fmt.Errorf("send request error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		maxSize := 4 * 1024
		limitedReader := io.LimitReader(resp.Body, int64(maxSize))
		data, err := io.ReadAll(limitedReader)
		if err != nil {
			return domain.RiskProfile{}, fmt.Errorf("read limited body error: %w", err)
		}

		return domain.RiskProfile{}, fmt.Errorf("status %d, data: %s", resp.StatusCode, string(data))
	}

	var profile domain.RiskProfile
	err = json.NewDecoder(resp.Body).Decode(&profile)
	if err != nil {
		return domain.RiskProfile{}, fmt.Errorf("parse response body error: %w", err)
	}

	return profile, nil
}
