package imports

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
	wakatime "github.com/muety/wakapi/models/compat/wakatime/v1"
	"github.com/muety/wakapi/utils"
	"net/http"
	"time"
)

// https://wakatime.com/api/v1/users/current/machine_names
// https://pastr.de/p/v58cv0xrupp3zvyyv8o6973j
func fetchMachineNames(baseUrl, apiKey string) (map[string]*wakatime.MachineEntry, error) {
	httpClient := &http.Client{Timeout: 10 * time.Second}

	machines := make(map[string]*wakatime.MachineEntry)

	for page := 1; ; page++ {
		url := fmt.Sprintf("%s%s?page=%d", baseUrl, config.WakatimeApiMachineNamesUrl, page)
		req, err := http.NewRequest(http.MethodGet, url, nil)
		req.Header.Set("Authorization", fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(apiKey))))
		if err != nil {
			return nil, err
		}

		res, err := httpClient.Do(req)
		if err != nil {
			return nil, err
		}
		defer res.Body.Close()

		var machineData wakatime.MachineViewModel
		if err := json.NewDecoder(res.Body).Decode(&machineData); err != nil {
			return nil, err
		}

		for _, ma := range machineData.Data {
			machines[ma.Id] = ma
		}

		if page == machineData.TotalPages {
			break
		}
	}

	return machines, nil
}

// https://wakatime.com/api/v1/users/current/user_agents
// https://pastr.de/p/05k5do8q108k94lic4lfl3pc
func fetchUserAgents(baseUrl, apiKey string) (map[string]*wakatime.UserAgentEntry, error) {
	httpClient := &http.Client{Timeout: 10 * time.Second}

	userAgents := make(map[string]*wakatime.UserAgentEntry)

	for page := 1; ; page++ {
		url := fmt.Sprintf("%s%s?page=%d", baseUrl, config.WakatimeApiUserAgentsUrl, page)
		req, err := http.NewRequest(http.MethodGet, url, nil)
		req.Header.Set("Authorization", fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(apiKey))))
		if err != nil {
			return nil, err
		}

		res, err := httpClient.Do(req)
		if err != nil {
			return nil, err
		}
		defer res.Body.Close()

		var userAgentsData wakatime.UserAgentsViewModel
		if err := json.NewDecoder(res.Body).Decode(&userAgentsData); err != nil {
			return nil, err
		}

		for _, ua := range userAgentsData.Data {
			userAgents[ua.Id] = ua
		}

		if page == userAgentsData.TotalPages {
			break
		}
	}

	return userAgents, nil
}

func mapHeartbeat(
	entry *wakatime.HeartbeatEntry,
	userAgents map[string]*wakatime.UserAgentEntry,
	machineNames map[string]*wakatime.MachineEntry,
	user *models.User,
) *models.Heartbeat {
	ua := userAgents[entry.UserAgentId]
	if ua == nil {
		// try to parse id as an actual user agent string (as returned by wakapi)
		if opSys, editor, err := utils.ParseUserAgent(entry.UserAgentId); err == nil {
			ua = &wakatime.UserAgentEntry{
				Editor: editor,
				Os:     opSys,
			}
		} else {
			ua = &wakatime.UserAgentEntry{
				Editor: "unknown",
				Os:     "unknown",
			}
		}
	}

	ma := machineNames[entry.MachineNameId]
	if ma == nil {
		ma = &wakatime.MachineEntry{
			Id:    entry.MachineNameId,
			Value: entry.MachineNameId,
		}
	}

	return (&models.Heartbeat{
		User:            user,
		UserID:          user.ID,
		Entity:          entry.Entity,
		Type:            entry.Type,
		Category:        entry.Category,
		Project:         entry.Project,
		Branch:          entry.Branch,
		Language:        entry.Language,
		IsWrite:         entry.IsWrite,
		Editor:          ua.Editor,
		OperatingSystem: ua.Os,
		Machine:         ma.Value,
		UserAgent:       ua.Value,
		Time:            models.CustomTime(time.Unix(0, int64(entry.Time*1e9))),
		Origin:          OriginWakatime,
		OriginId:        entry.Id,
		CreatedAt:       models.CustomTime(entry.CreatedAt),
	}).Hashed()
}
