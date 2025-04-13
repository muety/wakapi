package services

import (
	"github.com/muety/wakapi/mocks"
	"github.com/stretchr/testify/mock"
)

type ServicesMock struct {
	mock.Mock

	// Function overrides
	AliasFunc           func() IAliasService
	ProjectLabelFunc    func() IProjectLabelService
	DurationFunc        func() IDurationService
	SummaryFunc         func() ISummaryService
	KeyValueFunc        func() IKeyValueService
	HeartbeatFunc       func() IHeartbeatService
	UsersFunc           func() IUserService
	ActivityFunc        func() IActivityService
	AggregationFunc     func() IAggregationService
	ClientFunc          func() IClientService
	DiagnosticsFunc     func() IDiagnosticsService
	GoalFunc            func() IGoalService
	HouseKeepingFunc    func() IHousekeepingService
	InvoiceFunc         func() IInvoiceService
	LanguageMappingFunc func() ILanguageMappingService
	LeaderBoardFunc     func() ILeaderboardService
	MiscFunc            func() IMiscService
	OAuthFunc           func() IUserOauthService
	OtpFunc             func() IOTPService
	ReportFunc          func() IReportService
	UserAgentPluginFunc func() IPluginUserAgentService
}

// Default implementations
func (m *ServicesMock) Alias() IAliasService {
	if m.AliasFunc != nil {
		return m.AliasFunc()
	}
	return new(mocks.AliasServiceMock)
}

func (m *ServicesMock) ProjectLabel() IProjectLabelService {
	if m.ProjectLabelFunc != nil {
		return m.ProjectLabelFunc()
	}
	return new(mocks.ProjectLabelServiceMock)
}

func (m *ServicesMock) Duration() IDurationService {
	if m.DurationFunc != nil {
		return m.DurationFunc()
	}
	return new(mocks.DurationServiceMock)
}

func (m *ServicesMock) Summary() ISummaryService {
	if m.SummaryFunc != nil {
		return m.SummaryFunc()
	}
	return new(mocks.SummaryServiceMock)
}

func (m *ServicesMock) KeyValue() IKeyValueService {
	if m.KeyValueFunc != nil {
		return m.KeyValueFunc()
	}
	return new(mocks.KeyValueServiceMock)
}

func (m *ServicesMock) Heartbeat() IHeartbeatService {
	if m.HeartbeatFunc != nil {
		return m.HeartbeatFunc()
	}
	return new(mocks.HeartbeatServiceMock)
}

func (m *ServicesMock) Users() IUserService {
	if m.UsersFunc != nil {
		return m.UsersFunc()
	}
	return new(mocks.UserServiceMock)
}

// NOt Implemented
func (s *ServicesMock) Activity() IActivityService {
	return nil
}

func (s *ServicesMock) Aggregation() IAggregationService {
	return nil
}

func (s *ServicesMock) Client() IClientService {
	return nil
}

func (s *ServicesMock) Diagnostics() IDiagnosticsService {
	return nil
}

func (s *ServicesMock) Goal() IGoalService {
	return nil
}

func (s *ServicesMock) HouseKeeping() IHousekeepingService {
	return nil
}

func (s *ServicesMock) Invoice() IInvoiceService {
	return nil
}

func (s *ServicesMock) LanguageMapping() ILanguageMappingService {
	return nil
}

func (s *ServicesMock) LeaderBoard() ILeaderboardService {
	return nil
}

func (s *ServicesMock) Misc() IMiscService {
	return nil
}

func (s *ServicesMock) OAuth() IUserOauthService {
	return nil
}

func (s *ServicesMock) Otp() IOTPService {
	return nil
}

func (s *ServicesMock) Report() IReportService {
	return nil
}

func (s *ServicesMock) UserAgentPlugin() IPluginUserAgentService {
	return nil
}
