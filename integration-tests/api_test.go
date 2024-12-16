package integration_tests

import (
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"io"
	"net/http"
	"sphere-homework/app/config"
	"sphere-homework/app/dto"
	"strings"
)

var _ = Describe("Testing API endpoints", func() {
	_ = godotenv.Load("../.env")
	client := &http.Client{}
	cfg := config.NewConfig()
	baseUrl := fmt.Sprintf("http://localhost:%d/api/v1", cfg.Port)

	When("/transfer endpoint is invoked", func() {
		It("returns success for happy case", func() {
			request := dto.TransferRequest{
				FromAsset: "USD",
				ToAsset:   "GBP",
				Amount:    30000,
				Sender:    "jim",
				Recipient: "system",
			}

			b, err := json.Marshal(request)
			Expect(err).NotTo(HaveOccurred())
			reader := strings.NewReader(string(b))

			resp, err := client.Post(baseUrl+"/transfer", "application/json", reader)
			Expect(err).ToNot(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(http.StatusCreated))

			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			Expect(err).NotTo(HaveOccurred())

			response := dto.TransferResponse{}
			err = json.Unmarshal(body, &response)
			Expect(err).NotTo(HaveOccurred())
			Expect(response.TransferId).To(Not(BeNil()))
		})
	})
})
