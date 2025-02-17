package http

import (
	"encoding/json"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/LerianStudio/midaz/pkg"
	gold "github.com/LerianStudio/midaz/pkg/gold/transaction/model"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
)

type SimpleStruct struct {
	Name string
	Age  int
}

type ComplexStruct struct {
	Enable bool
	Simple SimpleStruct
}

func TestNewOfTypeWithSimpleStruct(t *testing.T) {
	s := newOfType(new(SimpleStruct))

	if err := json.Unmarshal([]byte("{\"Name\":\"Bruce\", \"Age\": 18}"), s); err != nil {
		t.Error(err)
	}

	sPrt := s.(*SimpleStruct)

	if sPrt.Name != "Bruce" || sPrt.Age != 18 {
		t.Error("Wrong data.")
	}
}

func TestNewOfTypeWithComplexStruct(t *testing.T) {
	s := newOfType(new(ComplexStruct))

	if err := json.Unmarshal([]byte("{\"Simple\": {\"Name\":\"Bruce\", \"Age\": 18}}"), s); err != nil {
		t.Error(err)
	}

	sPrt := s.(*ComplexStruct)

	if sPrt.Simple.Name != "Bruce" || sPrt.Simple.Age != 18 {
		t.Error("Wrong data.")
	}
}

func TestFilterRequiredFields(t *testing.T) {
	myMap := pkg.FieldValidations{
		"legalDocument":        "legalDocument is a required field",
		"legalName":            "legalName is a required field",
		"parentOrganizationId": "parentOrganizationId must be a valid UUID",
	}

	expected := pkg.FieldValidations{
		"legalDocument": "legalDocument is a required field",
		"legalName":     "legalName is a required field",
	}

	result := fieldsRequired(myMap)

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Want: %v, got %v", expected, result)
	}
}

func TestFilterRequiredFieldWithNoFields(t *testing.T) {
	myMap := pkg.FieldValidations{
		"parentOrganizationId": "parentOrganizationId must be a valid UUID",
	}

	expected := make(pkg.FieldValidations)
	result := fieldsRequired(myMap)

	if len(result) > 0 {
		t.Errorf("Want %v, got %v", expected, result)
	}
}

func TestParseUUIDPathParameters_ValidUUID(t *testing.T) {
	app := fiber.New()

	app.Get("/v1/organizations/:id", ParseUUIDPathParameters, func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK) // Se o middleware passar, responde com 200
	})

	req := httptest.NewRequest("GET", "/v1/organizations/123e4567-e89b-12d3-a456-426614174000", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	require.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestParseUUIDPathParameters_MultipleValidUUID(t *testing.T) {
	app := fiber.New()

	app.Get("/v1/organizations/:organization_id/ledgers/:id", ParseUUIDPathParameters, func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest(
		"GET",
		"/v1/organizations/123e4567-e89b-12d3-a456-426614174000/ledgers/c71ab589-cf46-4f2d-b6ef-b395c9a475da",
		nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	require.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestParseUUIDPathParameters_InvalidUUID(t *testing.T) {
	app := fiber.New()

	app.Get("/v1/organizations/:id", ParseUUIDPathParameters, func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest("GET", "/v1/organizations/invalid-uuid", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestParseUUIDPathParameters_ValidAndInvalidUUID(t *testing.T) {
	app := fiber.New()

	app.Get("/v1/organizations/:organization_id/ledgers/:id", ParseUUIDPathParameters, func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest(
		"GET",
		"/v1/organizations/123e4567-e89b-12d3-a456-426614174000/ledgers/invalid-uuid",
		nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestValidateStruct(t *testing.T) {
	t.Run("Valid Struct", func(t *testing.T) {
		type ValidStruct struct {
			Key   string `validate:"valuemax=10"`
			Value string `validate:"valuemax=20"`
		}

		validInput := ValidStruct{
			Key:   "validKey",
			Value: "validValue",
		}

		err := ValidateStruct(validInput)

		require.NoError(t, err, "expected no error for a valid struct")
	})

	t.Run("Required fields missing", func(t *testing.T) {
		type InvalidStruct struct {
			Name string `json:"name" validate:"required"`
		}

		invalidInput := InvalidStruct{
			Name: "",
		}
		err := ValidateStruct(invalidInput)

		require.Error(t, err, "expected an error warning required fields are missing")
	})

	t.Run("Key length exceeded", func(t *testing.T) {
		type InvalidStruct struct {
			Key   string `validate:"keymax=5"`
			Value string
		}
	
		invalidInput := InvalidStruct{
			Key:   "exceedingKeyLength",
			Value: "validValue",
		}
	
		err := ValidateStruct(invalidInput)
	
		require.Error(t, err, "expected an error for exceeding key length")
	})

	t.Run("Value length exceeded", func(t *testing.T) {
		type InvalidStruct struct {
			Key   string
			Value string `validate:"valuemax=20"`
		}
	
		invalidInput := InvalidStruct{
			Key:   "valid",
			Value: "exceedingValueMaxCharsLength",
		}
	
		err := ValidateStruct(invalidInput)
	
		require.Error(t, err, "expected an error for exceeding value length")
	})

	t.Run("No Nested field is nested", func(t *testing.T) {
		type InvalidStruct struct {
			Name     string         `json:"name"`
			Metadata map[string]any `json:"metadata" validate:"nonested"`
		}
	
		invalidInput := InvalidStruct{
			Name: "valid name",
			Metadata: map[string]any{
				"value1": "Value 1",
				"value2": "Value 2",
				"value3": "Value 3",
			},
		}
	
		err := ValidateStruct(invalidInput)
	
		require.Error(t, err, "expected an error for nesting a map")
	})

	t.Run("Transacation with more than one FromTo", func(t *testing.T) {
		type InvalidStruct struct {
			Amount int64         `json:"amount"`
			To     []gold.FromTo `json:"to" validate:"singletransactiontype"`
		}
	
		validFromTo := makeFromTo().
			addAmount(200).
			FromTo
	
		invalidFromTo := makeFromTo().
			addAmount(200).
			addShare(50).
			addRemaining("test").
			FromTo
	
		invalidInput := InvalidStruct{
			Amount: 200.0,
			To: []gold.FromTo{
				validFromTo,
				invalidFromTo,
			},
		}
	
		err := ValidateStruct(invalidInput)
	
		require.Error(t, err, "expected an error for transaction containing more than one transaction type (Amount, Share, Remaining)")
	})
}

type FromToExtension struct {
	FromTo gold.FromTo
}

func makeFromTo() *FromToExtension {
	return &FromToExtension{
		FromTo: gold.FromTo{
			Account:         "account",
			Rate:            nil,
			Description:     "description",
			ChartOfAccounts: "chart",
			Metadata:        nil,
			IsFrom:          true,
		},
	}
}

func (ftx *FromToExtension) addAmount(value int64) *FromToExtension {
	ftx.FromTo.Amount = &gold.Amount{
		Asset: "BRL",
		Value: value,
		Scale: 2,
	}

	return ftx
}

func (ftx *FromToExtension) addShare(value int64) *FromToExtension {
	ftx.FromTo.Share = &gold.Share{
		Percentage:             value,
		PercentageOfPercentage: value,
	}

	return ftx
}

func (ftx *FromToExtension) addRemaining(value string) *FromToExtension {
	ftx.FromTo.Remaining = value

	return ftx
}
