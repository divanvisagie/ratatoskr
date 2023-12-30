package layers

import (
	"math"
	"ratatoskr/pkg/client"
	"ratatoskr/pkg/repos"
	"ratatoskr/pkg/types"
	"ratatoskr/pkg/utils"
	"sort"
)

type MemoryLayer struct {
	child  Layer
	repo   *repos.MessageRepo
	logger *utils.Logger
}

func NewMemoryLayer(repo *repos.MessageRepo, child Layer) *MemoryLayer {
	logger := utils.NewLogger("Memory Layer")
	return &MemoryLayer{
		child,
		repo,
		logger,
	}
}

func cosineSimilarity(a, b []float32) float32 {
	var dotProduct, magnitudeA, magnitudeB float32
	for i := range a {
		dotProduct += a[i] * b[i]
		magnitudeA += a[i] * a[i]
		magnitudeB += b[i] * b[i]
	}
	return dotProduct / float32(math.Sqrt(float64(magnitudeA))*math.Sqrt(float64(magnitudeB)))
}

func (m *MemoryLayer) getContext(username string, inputMessage string) ([]types.StoredMessage, error) {
	var history []types.StoredMessage
	// Get the context in order to feed future prompts
	todaysMessages, err := m.repo.GetMessages(username)
	if err != nil {
		m.logger.Error("getContext, GetMessages", err)
		return nil, err
	}

	err, contextualMessages := m.repo.GetAllMessagesForUser(username)
	// take the top three embeddings that have cosine similarity
	msgEmbedding := client.Embed(inputMessage)
	for i := range contextualMessages {
		rank := cosineSimilarity(contextualMessages[i].Embedding, msgEmbedding)
		m.logger.Info("Calculated rank at %v\n", rank)
		contextualMessages[i].Rank = rank
	}

	// sort contextualmessages ranked highest to lowest
	sort.Slice(contextualMessages, func(i, j int) bool {
		return contextualMessages[i].Rank > contextualMessages[j].Rank
	})

	//append the top messages to the history
	for i := 0; i < 10; i++ {
		if i < len(contextualMessages) {
			history = append(history, contextualMessages[i])
		}
	}

	// Select only the latest 10 messages
	if len(history) > 10 {
		todaysMessages = todaysMessages[len(todaysMessages)-10:]
		history = append(history, todaysMessages...)
	}

	for _, ctxMsg := range history {
		m.logger.Info("Context message: %v: %v\n", ctxMsg.Role, ctxMsg.Message)
	}

	return history, nil
}

func (m *MemoryLayer) PassThrough(req *types.RequestMessage) (types.ResponseMessage, error) {
	inputMessage := repos.NewStoredMessage(repos.User, req.Message)
	m.repo.SaveMessage(req.UserName, *inputMessage)

	history, err := m.getContext(req.UserName, req.Message)
	if err != nil {
		m.logger.Error("PassThrough, getContext", err)
		return types.ResponseMessage{}, err
	}
	req.Context = history

	// Now pass through child layer
	res, err := m.child.PassThrough(req)
	if err != nil {
		m.logger.Error("PassThrough, child layer", err)
		return types.ResponseMessage{}, err
	}

	outputMessage := repos.NewStoredMessage(repos.Assistant, res.Message)
	m.repo.SaveMessage(req.UserName, *outputMessage)

	return res, nil
}
