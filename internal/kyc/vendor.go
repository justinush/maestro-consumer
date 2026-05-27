package kyc

import "fmt"

func FakeVendorCheckLiveness(applicantID string) error {
	fmt.Printf("vendor: liveness queued for applicant %s (stub)\n", applicantID)
	return nil
}
