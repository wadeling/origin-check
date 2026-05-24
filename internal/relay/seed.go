package relay

import (
	"context"
	"os"

	"github.com/wadeling/origin-check/internal/config"
	"github.com/wadeling/origin-check/internal/crypto"
	"github.com/wadeling/origin-check/internal/store"
	"gopkg.in/yaml.v3"
)

type SeedRelay struct {
	Name               string   `yaml:"name"`
	WebsiteURL         string   `yaml:"website_url"`
	APIBaseURL         string   `yaml:"api_base_url"`
	BackupAPIBaseURLs  []string `yaml:"backup_api_base_urls"`
	AuthType           string   `yaml:"auth_type"`
	APIKeyEnv          string   `yaml:"api_key_env"`
	ClaimedModels      []string `yaml:"claimed_models"`
	HealthModel        string   `yaml:"health_model"`
	Tags               []string `yaml:"tags"`
}

type SeedFile struct {
	Relays []SeedRelay `yaml:"relays"`
}

func DefaultSeeds() SeedFile {
	return SeedFile{
		Relays: []SeedRelay{
			{
				Name:              "Liaobots",
				WebsiteURL:        "https://liaobots.work/zh",
				APIBaseURL:        "https://ai.liaobots.work/v1",
				BackupAPIBaseURLs: []string{"https://ai.liaobots1.work/v1", "https://ai.liaobots2.work/v1"},
				AuthType:          "login_code",
				APIKeyEnv:         "LIAOBOTS_API_KEY",
				ClaimedModels:     append([]string(nil), MainstreamProbeModels...),
				HealthModel:       DefaultHealthModel,
				Tags:              []string{"聚合网关"},
			},
			{
				Name:          "灵芽 API",
				WebsiteURL:    "https://api.lingyaai.cn/",
				APIBaseURL:    "https://api.lingyaai.cn/v1",
				AuthType:      "bearer_token",
				APIKeyEnv:     "LINGYAAI_API_KEY",
				ClaimedModels: append([]string(nil), MainstreamProbeModels...),
				HealthModel:   DefaultHealthModel,
				Tags:          []string{"One-API"},
			},
			{
				Name:          "Asiai Cloud",
				WebsiteURL:    "https://api.asiai.cloud/",
				APIBaseURL:    "https://api.asiai.cloud/v1",
				AuthType:      "bearer_token",
				APIKeyEnv:     "ASIAI_API_KEY",
				ClaimedModels: append([]string(nil), MainstreamProbeModels...),
				HealthModel:   DefaultHealthModel,
				Tags:          []string{"待验证"},
			},
			{
				Name:          "NiceGoal",
				WebsiteURL:    "https://www.nicegoal.ai",
				APIBaseURL:    "https://www.nicegoal.ai/v1",
				AuthType:      "bearer_token",
				APIKeyEnv:     "NICEGOAL_API_KEY",
				ClaimedModels: append([]string(nil), MainstreamProbeModels...),
				HealthModel:   DefaultHealthModel,
				Tags:          []string{"New API"},
			},
		},
	}
}

func LoadSeedFile(path string) (SeedFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return SeedFile{}, err
	}
	var sf SeedFile
	if err := yaml.Unmarshal(data, &sf); err != nil {
		return SeedFile{}, err
	}
	return sf, nil
}

func Seed(ctx context.Context, st *store.Store, enc *crypto.Encryptor, sf SeedFile) error {
	for _, seed := range sf.Relays {
		var keyEncrypted []byte
		if seed.APIKeyEnv != "" {
			if key := config.Env(seed.APIKeyEnv); key != "" {
				encKey, err := enc.Encrypt(key)
				if err != nil {
					return err
				}
				keyEncrypted = encKey
			}
		}

		healthModel := seed.HealthModel
		if healthModel == "" {
			healthModel = DefaultHealthModel
		}

		r := &store.Relay{
			Name:              seed.Name,
			WebsiteURL:        seed.WebsiteURL,
			APIBaseURL:        seed.APIBaseURL,
			BackupAPIBaseURLs: seed.BackupAPIBaseURLs,
			APIKeyEncrypted:   keyEncrypted,
			AuthType:          seed.AuthType,
			ClaimedModels:     seed.ClaimedModels,
			HealthModel:       healthModel,
			Status:            store.RelayActive,
			Tags:              seed.Tags,
		}
		if r.BackupAPIBaseURLs == nil {
			r.BackupAPIBaseURLs = []string{}
		}
		if r.ClaimedModels == nil {
			r.ClaimedModels = []string{}
		}
		if r.Tags == nil {
			r.Tags = []string{}
		}
		if err := st.UpsertRelay(ctx, r); err != nil {
			return err
		}
	}
	return nil
}
