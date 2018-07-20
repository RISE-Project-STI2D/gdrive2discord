package gdrive2discord

import (
	"encoding/json"
	"net/http"
	"regexp"

	"../discord"
	"../google"
	"../google/userinfo"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
)

type Request struct {
	GoogleCode        string   `json:"g"`
	DiscordWebhookURL string   `json:"w"`
	FolderIds         []string `json:"fids"`
	FolderName        string   `json:"fn"`
}

type ErrResponse struct {
	Error string `json:"error"`
}

func ServeHttp(env *Environment) {
	r := martini.NewRouter()
	mr := martini.New()
	mr.Use(martini.Recovery())
	mr.Use(martini.Static("public", martini.StaticOptions{
		SkipLogging: true,
	}))
	mr.MapTo(r, (*martini.Routes)(nil))
	mr.Action(r.Handle)
	m := &martini.ClassicMartini{mr, r}
	m.Use(render.Renderer())

	m.Get("/", func(renderer render.Render, req *http.Request) {
		renderer.HTML(200, "index", env)
	})
	m.Put("/", func(renderer render.Render, req *http.Request) {
		handleSubscriptionRequest(env, renderer, req)
	})
	m.RunOnAddr(env.Configuration.BindAddress)
}

func handleSubscriptionRequest(env *Environment, renderer render.Render, req *http.Request) {
	decoder := json.NewDecoder(req.Body)
	var r Request
	err := decoder.Decode(&r)
	if err != nil {
		renderer.JSON(400, &ErrResponse{err.Error()})
		return
	}
	if r.GoogleCode == "" {
		renderer.JSON(400, &ErrResponse{"Invalid oauth code for google"})
		return
	}

	if r.DiscordWebhookURL == "" {
		renderer.JSON(400, &ErrResponse{"Invalid webhook for discord"})
		return
	}
	matched, err := regexp.MatchString("https?://discordapp.com/api/webhooks/[0-9]+/.*", r.DiscordWebhookURL)
	if !matched {
		renderer.JSON(400, &ErrResponse{"Invalid webhook for discord"})
		return
	}
	googleRefreshToken, googleAccessToken, status, err := google.NewAccessToken(env.Configuration.Google, env.HttpClient, r.GoogleCode)
	if status != google.Ok {
		renderer.JSON(500, &ErrResponse{err.Error()})
		return
	}
	gUserInfo, status, err := userinfo.GetUserInfo(env.HttpClient, googleAccessToken)
	if status != google.Ok {
		renderer.JSON(500, &ErrResponse{err.Error()})
		return
	}
	webhookInfo, sstatus, err := discord.GetWebhookInfo(env.HttpClient, r.DiscordWebhookURL)
	if sstatus != discord.Ok {
		renderer.JSON(500, &ErrResponse{err.Error()})
		return
	}
	welcomeMessage := CreateDiscordWelcomeMessage(env.Configuration.Google.RedirectUri, gUserInfo, env.Version)
	cstatus, err := discord.PostMessage(env.HttpClient, r.DiscordWebhookURL, welcomeMessage)

	env.RegisterChannel <- &SubscriptionAndAccessToken{
		Subscription: &Subscription{
			r.DiscordWebhookURL,
			googleRefreshToken,
			gUserInfo,
			webhookInfo,
			r.FolderIds,
		},
		GoogleAccessToken: googleAccessToken,
	}

	renderer.JSON(200, map[string]interface{}{
		"user":         gUserInfo,
		"channelFound": cstatus == discord.Ok,
	})

}
