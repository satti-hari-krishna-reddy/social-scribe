package services

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/dghubble/oauth1"
	"social-scribe/backend/internal/models"
	"social-scribe/backend/internal/repositories"
)

func ProcessSharedBlog(user *models.User, blogId string, platforms []string) error {
	userId := user.Id.Hex()

	if !user.Verified {
		return fmt.Errorf("user is not verified")
	}
	validPlatforms := map[string]bool{
		"twitter":  true,
		"linkedin": true,
	}
	if len(platforms) == 0 {
		return fmt.Errorf("at least one platform must be specified")
	}
	for _, platform := range platforms {
		if !validPlatforms[platform] {
			return fmt.Errorf("invalid platform specified")
		}
	}
	query := models.GraphQLQuery{
		Query: `query Post($id: ID!) {
            post(id: $id) {
                id
                url
                coverImage {
                    url
                }
                author {
                    name
                }
                readTimeInMinutes
                title
                subtitle
                brief
                content {
                    text
                }
            }
        }`,
		Variables: map[string]interface{}{
			"id": blogId,
		},
	}
	queryBytes, err := json.Marshal(query)
	if err != nil {
		return fmt.Errorf("failed to marshal query: %v", err)
	}
	endpoint := "https://gql.hashnode.com"
	headers := map[string]string{"Content-Type": "application/json"}
	gqlResponse, err := MakePostRequest(endpoint, queryBytes, headers)
	if err != nil {
		return fmt.Errorf("failed to make request: %v", err)
	}
	var response struct {
		Data struct {
			Post struct {
				Id         string `json:"id"`
				Title      string `json:"title"`
				Url        string `json:"url"`
				CoverImage struct {
					Url string `json:"url"`
				} `json:"coverImage"`
				Author struct {
					Name string `json:"name"`
				} `json:"author"`
				ReadTimeInMinutes int    `json:"readTimeInMinutes"`
				SubTitle          string `json:"subtitle"`
				Brief             string `json:"brief"`
				Content           struct {
					Text string `json:"text"`
				} `json:"content"`
			} `json:"post"`
		} `json:"data"`
	}
	if err := json.Unmarshal(gqlResponse, &response); err != nil {
		return fmt.Errorf("failed to unmarshal response: %v", err)
	}
	const maxContentLength = 150
	content := response.Data.Post.Content.Text
	if len(content) > maxContentLength {
		content = content[:maxContentLength] + "..."
	}
	prompt := fmt.Sprintf(
		"Write a post that i can post it in linkedin and twitter (X) for this blog:\n\n"+
			"Title: %s\n"+
			"Subtitle: %s\n"+
			"Brief: %s\n"+
			"Content snippet: %s\n\n"+
			"Note: The tone should be human, engaging, and conversational. Encourage readers to click the blog link for more details. Avoid sounding robotic or generic. Mention the blog’s key takeaway and invite readers to check it out and make sure it was short enough and dont be too verbose as twitter and linkedin has character limit on how much we can tweet or post so please keep it short and also make sure to generate single post that can be used for both linkedin and twitter rather seperately and dont use any wild card characters like * and without commentary.",
		response.Data.Post.Title,
		response.Data.Post.SubTitle,
		response.Data.Post.Brief,
		content,
	)
	aiResponse, err := invokeAi(prompt)
	if err != nil {
		return fmt.Errorf("failed to generate post content: %v", err)
	}
	for _, platform := range platforms {
		switch platform {
		case "linkedin":
			err = linkedPostHandler(aiResponse, user.LinkedInOauthKey)
			if err != nil {
				return fmt.Errorf("failed to post content to LinkedIn: %v", err)
			}
		case "twitter":
			token := oauth1.NewToken(user.XOAuthToken, user.XOAuthSecret)
			err = postTweetHandler(aiResponse, blogId, token)
			if err != nil {
				return fmt.Errorf("failed to post content to Twitter: %v", err)
			}
		}
	}
	var isFound bool
	for i := range user.SharedBlogs {
		if user.SharedBlogs[i].Id == response.Data.Post.Id {
			user.SharedBlogs[i].SharedTime = time.Now().Format(time.RFC3339)
			err = repositories.UpdateUser(userId, user)
			isFound = true
			if err != nil {
				return fmt.Errorf("failed to update user with shared blog: %v", err)
			}
			break
		}
	}
	if !isFound {
		var newSharedBlog models.SharedBlog
		newSharedBlog.Id = response.Data.Post.Id
		newSharedBlog.Title = response.Data.Post.Title
		newSharedBlog.Url = response.Data.Post.Url
		newSharedBlog.CoverImage = models.Image{URL: response.Data.Post.CoverImage.Url}
		newSharedBlog.Author = models.Author{Name: response.Data.Post.Author.Name}
		newSharedBlog.ReadTimeInMinutes = response.Data.Post.ReadTimeInMinutes
		newSharedBlog.SharedTime = time.Now().Format(time.RFC3339)
		user.SharedBlogs = append(user.SharedBlogs, newSharedBlog)
		err = repositories.UpdateUser(userId, user)
		if err != nil {
			return fmt.Errorf("failed to update user with shared blog: %v", err)
		}
	}
	return nil
}
