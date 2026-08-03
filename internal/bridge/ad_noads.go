//go:build noads

package bridge

import "fmt"

type AdRuntime struct{}
type AdService struct{}

func NewAdService(...any) *AdService { return &AdService{} }
func (service *AdService) GetAdRuntime() (AdRuntime, error) {
	return AdRuntime{}, fmt.Errorf("广告功能已禁用")
}
func (service *AdService) OpenExternalURL(string) error {
	return fmt.Errorf("广告功能已禁用")
}
