package dashboard

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Summary() (*Summary, error) {
	return s.repo.Summary()
}

func (s *Service) MonthlyRevenue(months int) ([]MonthlyRevenue, error) {
	return s.repo.MonthlyRevenue(months)
}

func (s *Service) RevenueByCity() ([]CityRevenue, error) {
	return s.repo.RevenueByCity()
}

func (s *Service) RevenueByFeeType() ([]FeeTypeRevenue, error) {
	return s.repo.RevenueByFeeType()
}
