package orchestrator

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/looplj/axonhub/internal/ent/model"
	"github.com/looplj/axonhub/internal/objects"
	"github.com/looplj/axonhub/internal/server/biz"
	"github.com/looplj/axonhub/llm"
)

func TestSelectCandidates_APIKeyModelAssociationRequiresEnabledModel(t *testing.T) {
	ctx, client := setupTest(t)

	_, err := client.Model.Create().
		SetModelID("gpt-4-disabled").
		SetName("Disabled GPT-4").
		SetDeveloper("openai").
		SetIcon("openai").
		SetGroup("gpt-4").
		SetModelCard(&objects.ModelCard{}).
		SetStatus(model.StatusDisabled).
		SetSettings(&objects.ModelSettings{}).
		Save(ctx)
	require.NoError(t, err)

	project := createTestProject(t, ctx, client)
	user, err := client.User.Create().
		SetEmail("api-key-policy-test@example.com").
		SetPassword("password").
		Save(ctx)
	require.NoError(t, err)

	apiKey, err := client.APIKey.Create().
		SetName("policy key").
		SetKey("ah-policy-key").
		SetProjectID(project.ID).
		SetUserID(user.ID).
		SetProfiles(&objects.APIKeyProfiles{
			ActiveProfile: "default",
			Profiles: []objects.APIKeyProfile{
				{
					Name:          "default",
					ModelMappings: []objects.ModelMapping{},
					ModelAssociations: []objects.APIKeyModelAssociationPolicy{
						{
							ModelID: "gpt-4-disabled",
							Associations: []*objects.ModelAssociation{
								{
									Type: "model",
									ModelID: &objects.ModelIDAssociation{
										ModelID: "gpt-4",
									},
								},
							},
						},
					},
				},
			},
		}).
		Save(ctx)
	require.NoError(t, err)

	systemService := newTestSystemService(client)
	inbound := &PersistentInboundTransformer{
		state: &PersistenceState{
			APIKey:            apiKey,
			ModelService:      newTestModelService(client),
			CandidateSelector: &staticChannelSelector{},
		},
	}

	_, err = selectCandidates(inbound, nil, systemService).OnInboundLlmRequest(ctx, &llm.Request{Model: "gpt-4-disabled"})
	require.True(t, errors.Is(err, biz.ErrInvalidModel), "disabled models must not be revived by API key model associations")
}
