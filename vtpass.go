package vtupass_go

import (
	"context"
	"encoding/json"
	"fmt"

	// "fmt"
	"net/http"

	httpclient "github.com/CeoFred/vtupass_go/lib"
)

type VTService struct {
	apiKey          string
	publicKey       string
	secretKey       string
	client          HttpClient
	authCredentials map[string]string
	Enviroment      Environment
}

type BaseResponse struct {
	Code string `json:"code"`
}

type ErrorResponse struct {
	BaseResponse
}

func (e ErrorResponse) Error() string {

	var message string

	switch e.Code {
	case BILLER_CONFIRMED:
		message = "BILLER CONFIRMED"
	case INVALID_ARGUMENTS:
		message = "INVALID ARGUMENTS"
	case PRODUCT_DOES_NOT_EXIST:
		message = "PRODUCT_DOES_NOT_EXIST"
	case BILLER_NOT_REACHABLE_AT_THIS_POINT:
		message = "BILLER NOT REACHABLE AT THIS POINT"
	}
	return message
}

type WalletBalance struct {
	BaseResponse
	Contents struct {
		Balance float64 `json:"balance"`
	} `json:"contents"`
}

type ServiceCategoryResponse struct {
	BaseResponse
	Content             []ServiceCategory `json:"content"`
	ResponseDescription string            `json:"response_description"`
}

type ServiceResponse struct {
	BaseResponse
	Content             []Service `json:"content"`
	ResponseDescription string    `json:"response_description"`
}

type VariationResponse struct {
	BaseResponse
	Content struct {
		ServiceName string      `json:"ServiceName"`
		Variations  []Variation `json:"varations"`
	} `json:"content"`
}

type CustomerInfoResponse struct {
	BaseResponse
	Content CustomerInfo `json:"content"`
}

func NewVTService(apiKey, publicKey, secretKey string, environment Environment) *VTService {
	var baseUrl string

	switch environment {
	case EnvironmentSandbox:
		baseUrl = SandboxBaseURL
	case EnvironmentLive:
		baseUrl = LiveEnviromentURL
	default:
		baseUrl = SandboxBaseURL
	}

	return &VTService{
		apiKey:     apiKey,
		client:     httpclient.NewAPIClient(baseUrl, apiKey),
		Enviroment: environment,
		publicKey:  publicKey,
		secretKey:  secretKey,
		authCredentials: map[string]string{
			"api-key":    apiKey,
			"public-key": publicKey,
			"secret-key": secretKey,
		},
	}
}

//VERIFY METER NUMBER
// https://www.vtpass.com/documentation/eedc-enugu-electric-api/
func (s *VTService) VerifyMeterNumber(ctx context.Context, meter_number, meter_type, service_id string) (*CustomerInfo, error) {
	url := "merchant-verify"

	requestData := map[string]interface{}{
		"billersCode": meter_number,
		"serviceID":   service_id,
		"type":        meter_type,
	}

	resp, err := s.client.Post(ctx, url, requestData, s.authCredentials)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errorResponse ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errorResponse); err != nil {
			return nil, err
		}

		return nil, errorResponse
	}

	var resonse CustomerInfoResponse
	if err := json.NewDecoder(resp.Body).Decode(&resonse); err != nil {
		return nil, err
	}

	fmt.Println(resonse)

	if resonse.Code == "011" {
		return nil, fmt.Errorf("service not valid or invalid argumments")
	}
	if resonse.Code == "012" {
		return nil, fmt.Errorf("prodduct does not exist")
	}

	return &resonse.Content, nil

}

// GET VARIATION CODES
// https://www.vtpass.com/documentation/variation-codes/
func (s *VTService) ServiceVariations(ctx context.Context, id string) ([]Variation, error) {
	url := fmt.Sprintf("service-variations?serviceID=%s", id)

	resp, err := s.client.Get(ctx, url, s.authCredentials)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errorResponse ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errorResponse); err != nil {
			return nil, err
		}

		return nil, errorResponse
	}
	var resonse VariationResponse
	if err := json.NewDecoder(resp.Body).Decode(&resonse); err != nil {
		return nil, err
	}

	if resonse.Code == "011" {
		return nil, fmt.Errorf("service not valid or invalid argumments")
	}

	return resonse.Content.Variations, nil
}

// GET SERVICE ID
// https://www.vtpass.com/documentation/service-ids/
func (s *VTService) ServiceByIdentifier(ctx context.Context, id string) ([]Service, error) {
	url := fmt.Sprintf("services?identifier=%s", id)

	resp, err := s.client.Get(ctx, url, s.authCredentials)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errorResponse ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errorResponse); err != nil {
			return nil, err
		}

		return nil, errorResponse
	}
	var resonse ServiceResponse
	if err := json.NewDecoder(resp.Body).Decode(&resonse); err != nil {
		return nil, err
	}
	if resonse.Code == "011" {
		return nil, fmt.Errorf("service not valid or invalid argumments")
	}

	return resonse.Content, nil
}

func (s *VTService) ServiceCategories(ctx context.Context) ([]ServiceCategory, error) {
	resp, err := s.client.Get(ctx, "service-categories", s.authCredentials)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errorResponse ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errorResponse); err != nil {
			return nil, err
		}

		return nil, errorResponse
	}
	var resonse ServiceCategoryResponse
	if err := json.NewDecoder(resp.Body).Decode(&resonse); err != nil {
		return nil, err
	}
	if resonse.Code == "011" {
		return nil, fmt.Errorf("service not valid or invalid argumments")
	}

	return resonse.Content, nil
}

// Test authentication
func (s *VTService) Ping(ctx context.Context) (bool, error) {
	resp, err := s.client.Get(ctx, "balance", s.authCredentials)
	if err != nil {
		return false, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errorResponse ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errorResponse); err != nil {
			return false, err
		}

		return false, errorResponse
	}
	return true, nil
}

func (s *VTService) Balance(ctx context.Context) (*WalletBalance, error) {

	resp, err := s.client.Get(ctx, "balance", s.authCredentials)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errorResponse ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errorResponse); err != nil {
			return nil, err
		}

		return nil, errorResponse
	}

	var resonse WalletBalance
	if err := json.NewDecoder(resp.Body).Decode(&resonse); err != nil {
		return nil, err
	}
	if resonse.Code == "011" {
		return nil, fmt.Errorf("service not valid or invalid argumments")
	}
	return &resonse, nil

}

// func (c *Service) PostData(ctx context.Context, payload interface{}) (*Response, error) {

// 	path := fmt.Sprintf("user%s", c.authCredentials)

// 	resp, err := c.client.Post(ctx, path, payload)

// 	if err != nil {
// 		fmt.Println(err.Error())
// 		return nil, err
// 	}

// 	defer resp.Body.Close()

// 	if resp.StatusCode != http.StatusOK {
// 		var errorResponse ErrorResponse
// 		if err := json.NewDecoder(resp.Body).Decode(&errorResponse); err != nil {
// 			return nil, err
// 		}

// 		return nil, errorResponse
// 	}

// 	var response Response
// 	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
// 		fmt.Println("error decoding response", err)
// 		return nil, err
// 	}

// 	return &response, nil
// }