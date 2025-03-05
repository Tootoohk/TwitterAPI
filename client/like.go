package client

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Tootoohk/TwitterAPI/client/addons"
	"github.com/Tootoohk/TwitterAPI/models"
	"github.com/Tootoohk/TwitterAPI/utils"
)

// Like adds a like to a tweet
// tweetID can be either a tweet URL or tweet ID
func (t *Twitter) Like(tweetID string) *models.ActionResponse {
	// Extract tweet ID if URL was provided
	if strings.Contains(tweetID, "twitter.com") || strings.Contains(tweetID, "x.com") {
		var err error
		tweetID, err = addons.ExtractTweetID(tweetID, t.Account.Username, t.Logger)
		if err != nil {
			return &models.ActionResponse{
				Success: false,
				Error:   fmt.Errorf("invalid tweet URL: %w", err),
				Status:  models.StatusUnknown,
			}
		}
	}

	// Build URL and request body
	baseURL := "https://twitter.com/i/api/graphql/" + t.Config.Constants.QueryID.Like + "/FavoriteTweet"
	requestBody := fmt.Sprintf(`{"variables":{"tweet_id":"%s"},"queryId":"%s"}`,
		tweetID, t.Config.Constants.QueryID.Like)

	// Create request config
	reqConfig := utils.DefaultConfig()
	reqConfig.Method = "POST"
	reqConfig.URL = baseURL
	reqConfig.Body = strings.NewReader(requestBody)
	reqConfig.Headers = append(reqConfig.Headers,
		utils.HeaderPair{Key: "accept", Value: "*/*"},
		utils.HeaderPair{Key: "authorization", Value: t.Config.Constants.BearerToken},
		utils.HeaderPair{Key: "content-type", Value: "application/json"},
		utils.HeaderPair{Key: "cookie", Value: t.Cookies.CookiesToHeader()},
		utils.HeaderPair{Key: "origin", Value: "https://twitter.com"},
		utils.HeaderPair{Key: "x-csrf-token", Value: t.Account.Ct0},
		utils.HeaderPair{Key: "x-twitter-active-user", Value: "yes"},
		utils.HeaderPair{Key: "x-twitter-auth-type", Value: "OAuth2Session"},
	)

	// Make the request
	bodyBytes, resp, err := utils.MakeRequest(t.Client, reqConfig)
	if err != nil {
		t.Logger.Error("%s | Failed to like tweet %s: %v", t.Account.Username, tweetID, err)
		return &models.ActionResponse{
			Success: false,
			Error:   err,
			Status:  models.StatusUnknown,
		}
	}

	// Update cookies
	t.Cookies.SetCookieFromResponse(resp)
	if newCt0, ok := t.Cookies.GetCookieValue("ct0"); ok {
		t.Account.Ct0 = newCt0
	}

	bodyString := string(bodyBytes)

	// Handle successful responses
	if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
		if strings.Contains(bodyString, "already") {
			t.Logger.Success("%s | Tweet %s was already liked", t.Account.Username, tweetID)
			return &models.ActionResponse{
				Success: true,
				Status:  models.StatusAlreadyDone,
			}
		}

		var response models.LikeGraphQLResponse
		if err := json.Unmarshal(bodyBytes, &response); err != nil {
			t.Logger.Error("%s | Failed to parse like response for tweet %s: %v", t.Account.Username, tweetID, err)
			return &models.ActionResponse{
				Success: false,
				Error:   err,
				Status:  models.StatusUnknown,
			}
		}

		if response.Data.FavoriteTweet == "Done" {
			t.Logger.Success("%s | Successfully liked tweet %s", t.Account.Username, tweetID)
			return &models.ActionResponse{
				Success: true,
				Status:  models.StatusSuccess,
			}
		} else {
			t.Logger.Error("%s | Failed to like tweet %s: unexpected response", t.Account.Username, tweetID)
			return &models.ActionResponse{
				Success: false,
				Error:   fmt.Errorf("unexpected response: %s", response.Data.FavoriteTweet),
				Status:  models.StatusUnknown,
			}
		}
	}

	// Handle error responses
	switch {
	case strings.Contains(bodyString, "this account is temporarily locked"):
		t.Logger.Error("%s | Account is temporarily locked", t.Account.Username)
		return &models.ActionResponse{
			Success: false,
			Error:   models.ErrAccountLocked,
			Status:  models.StatusLocked,
		}
	case strings.Contains(bodyString, "Could not authenticate you"):
		t.Logger.Error("%s | Could not authenticate you", t.Account.Username)
		return &models.ActionResponse{
			Success: false,
			Error:   models.ErrAuthFailed,
			Status:  models.StatusAuthError,
		}
	default:
		t.Logger.Error("%s | Unknown response: %s", t.Account.Username, bodyString)
		return &models.ActionResponse{
			Success: false,
			Error:   fmt.Errorf("unknown response: %s", bodyString),
			Status:  models.StatusUnknown,
		}
	}
}
