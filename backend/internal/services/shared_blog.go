package services

import (
	"encoding/json"
	"fmt"
	"strings"
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
	// so the idea is to tell the ai to generate both posts in a single request
	prompt := fmt.Sprintf(
		"Generate two separate social media posts for this blog, one for Twitter (X) and one for LinkedIn.\n\n"+
			"Title: %s\n"+
			"Url: %s\n"+
			"Subtitle: %s\n"+
			"Brief: %s\n"+
			"Content snippet: %s\n\n"+
			"--- Output Format ---\n"+
			"[TWITTER]\n"+
			"Your Twitter post here (must be **280 characters or less**, including hashtags and URL)\n\n"+
			"[LINKEDIN]\n"+
			"Your LinkedIn post here (no strict length limit)\n\n"+
			"--- Additional Instructions ---\n"+
			"- Keep the tone **engaging, conversational, and human**.\n"+
			"- The Twitter post **MUST fit in 280 characters including hashtags & URL**.\n"+
			"- Ensure the blog URL is included in both posts.\n"+
			"- Do NOT add any extra commentary or explanations.\n"+
			"- Include relevant **hashtags** in both posts.\n"+
			"- Make the **LinkedIn post slightly longer**, but still concise and engaging.\n"+
			"- Ensure that the response format is **EXACTLY as specified**, so it can be parsed programmatically.\n",
		response.Data.Post.Title,
		response.Data.Post.Url,
		response.Data.Post.SubTitle,
		response.Data.Post.Brief,
		content,
	)

	aiResponse, err := invokeAi(prompt)
	if err != nil {
		return fmt.Errorf("failed to generate post content: %v", err)
	}

	// Splitting the response
	twitterTag := "[TWITTER]"
	linkedinTag := "[LINKEDIN]"

	twitterStart := strings.Index(aiResponse, twitterTag)
	linkedinStart := strings.Index(aiResponse, linkedinTag)

	// Extract Twitter content
	var twitterPost string
	if twitterStart != -1 && linkedinStart != -1 {
		twitterPost = strings.TrimSpace(aiResponse[twitterStart+len(twitterTag) : linkedinStart])
	} else if twitterStart != -1 {
		twitterPost = strings.TrimSpace(aiResponse[twitterStart+len(twitterTag):])
	}

	// Extract LinkedIn content
	var linkedinPost string
	if linkedinStart != -1 {
		linkedinPost = strings.TrimSpace(aiResponse[linkedinStart+len(linkedinTag):])
	}

	for _, platform := range platforms {
		switch platform {
		case "linkedin":
			err = linkedPostHandler(linkedinPost, user.LinkedInOauthKey)
			if err != nil {
				return fmt.Errorf("failed to post content to LinkedIn: %v", err)
			}
		case "twitter":
			token := oauth1.NewToken(user.XOAuthToken, user.XOAuthSecret)
			err = postTweetHandler(twitterPost, blogId, token)
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
