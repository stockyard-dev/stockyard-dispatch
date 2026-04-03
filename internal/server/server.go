package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/smtp"
	"strings"
	"time"

	"github.com/stockyard-dev/stockyard-dispatch/internal/store"
)

type SMTPConfig struct {
	Host string
	Port string
	User string
	Pass string
	From string
}

type Server struct {
	db     *store.DB
	mux    *http.ServeMux
	port   int
	limits Limits
	smtp   SMTPConfig
}

func New(db *store.DB, port int, limits Limits, smtpCfg SMTPConfig) *Server {
	s := &Server{db: db, mux: http.NewServeMux(), port: port, limits: limits, smtp: smtpCfg}
	s.routes()
	return s
}

func (s *Server) routes() {
	// Lists
	s.mux.HandleFunc("POST /api/lists", s.handleCreateList)
	s.mux.HandleFunc("GET /api/lists", s.handleListLists)
	s.mux.HandleFunc("GET /api/lists/{id}", s.handleGetList)
	s.mux.HandleFunc("DELETE /api/lists/{id}", s.handleDeleteList)

	// Subscribers
	s.mux.HandleFunc("POST /api/lists/{id}/subscribers", s.handleAddSubscriber)
	s.mux.HandleFunc("GET /api/lists/{id}/subscribers", s.handleListSubscribers)
	s.mux.HandleFunc("DELETE /api/subscribers/{id}", s.handleDeleteSubscriber)

	// Public subscribe/unsubscribe
	s.mux.HandleFunc("POST /subscribe/{id}", s.handlePublicSubscribe)
	s.mux.HandleFunc("GET /unsubscribe", s.handleUnsubscribe)

	// Campaigns
	s.mux.HandleFunc("POST /api/lists/{id}/campaigns", s.handleCreateCampaign)
	s.mux.HandleFunc("GET /api/lists/{id}/campaigns", s.handleListCampaigns)
	s.mux.HandleFunc("GET /api/campaigns/{id}", s.handleGetCampaign)
	s.mux.HandleFunc("POST /api/campaigns/{id}/send", s.handleSendCampaign)
	s.mux.HandleFunc("DELETE /api/campaigns/{id}", s.handleDeleteCampaign)
	s.mux.HandleFunc("GET /api/campaigns/{id}/sends", s.handleListSends)

	// Tracking
	s.mux.HandleFunc("GET /track/open/{id}", s.handleTrackOpen)

	// Status
	s.mux.HandleFunc("GET /api/status", s.handleStatus)
	s.mux.HandleFunc("GET /health", s.handleHealth)
	s.mux.HandleFunc("GET /ui", s.handleUI)
	s.mux.HandleFunc("GET /api/version", func(w http.ResponseWriter, r *http.Request) {
s.mux.HandleFunc("GET /api/tier",func(w http.ResponseWriter,r *http.Request){writeJSON(w,200,map[string]any{"tier":s.limits.Tier,"upgrade_url":"https://stockyard.dev/dispatch/"})})
		writeJSON(w, 200, map[string]any{"product": "stockyard-dispatch", "version": "0.1.0"})
	})
}

func (s *Server) Start() error {
	addr := fmt.Sprintf(":%d", s.port)
	log.Printf("[dispatch] listening on %s", addr)
	return http.ListenAndServe(addr, s.mux)
}

// --- List handlers ---

func (s *Server) handleCreateList(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Name == "" {
		writeJSON(w, 400, map[string]string{"error": "name is required"})
		return
	}
	if s.limits.MaxLists > 0 {
		lists, _ := s.db.ListLists()
		if LimitReached(s.limits.MaxLists, len(lists)) {
			writeJSON(w, 402, map[string]string{"error": fmt.Sprintf("free tier limit: %d list(s) — upgrade to Pro", s.limits.MaxLists), "upgrade": "https://stockyard.dev/dispatch/"})
			return
		}
	}
	l, err := s.db.CreateList(req.Name, req.Description)
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, 201, map[string]any{"list": l})
}

func (s *Server) handleListLists(w http.ResponseWriter, r *http.Request) {
	lists, err := s.db.ListLists()
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	if lists == nil {
		lists = []store.List{}
	}
	writeJSON(w, 200, map[string]any{"lists": lists, "count": len(lists)})
}

func (s *Server) handleGetList(w http.ResponseWriter, r *http.Request) {
	l, err := s.db.GetList(r.PathValue("id"))
	if err != nil {
		writeJSON(w, 404, map[string]string{"error": "list not found"})
		return
	}
	writeJSON(w, 200, map[string]any{"list": l})
}

func (s *Server) handleDeleteList(w http.ResponseWriter, r *http.Request) {
	s.db.DeleteList(r.PathValue("id"))
	writeJSON(w, 200, map[string]string{"status": "deleted"})
}

// --- Subscriber handlers ---

func (s *Server) handleAddSubscriber(w http.ResponseWriter, r *http.Request) {
	listID := r.PathValue("id")
	if _, err := s.db.GetList(listID); err != nil {
		writeJSON(w, 404, map[string]string{"error": "list not found"})
		return
	}
	var req struct {
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Email == "" {
		writeJSON(w, 400, map[string]string{"error": "email is required"})
		return
	}
	if s.limits.MaxSubscribers > 0 {
		total := s.db.TotalSubscribers()
		if LimitReached(s.limits.MaxSubscribers, total) {
			writeJSON(w, 402, map[string]string{"error": fmt.Sprintf("free tier limit: %d subscribers — upgrade to Pro", s.limits.MaxSubscribers), "upgrade": "https://stockyard.dev/dispatch/"})
			return
		}
	}
	sub, err := s.db.AddSubscriber(listID, req.Email, req.Name)
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, 201, map[string]any{"subscriber": sub})
}

func (s *Server) handleListSubscribers(w http.ResponseWriter, r *http.Request) {
	subs, err := s.db.ListSubscribers(r.PathValue("id"), 200)
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	if subs == nil {
		subs = []store.Subscriber{}
	}
	writeJSON(w, 200, map[string]any{"subscribers": subs, "count": len(subs)})
}

func (s *Server) handleDeleteSubscriber(w http.ResponseWriter, r *http.Request) {
	s.db.DeleteSubscriber(r.PathValue("id"))
	writeJSON(w, 200, map[string]string{"status": "deleted"})
}

// --- Public subscribe/unsubscribe ---

func (s *Server) handlePublicSubscribe(w http.ResponseWriter, r *http.Request) {
	listID := r.PathValue("id")
	if _, err := s.db.GetList(listID); err != nil {
		http.Error(w, "List not found", 404)
		return
	}

	var email, name string
	ct := r.Header.Get("Content-Type")
	if strings.HasPrefix(ct, "application/json") {
		var req struct {
			Email string `json:"email"`
			Name  string `json:"name"`
		}
		json.NewDecoder(r.Body).Decode(&req)
		email = req.Email
		name = req.Name
	} else {
		r.ParseForm()
		email = r.FormValue("email")
		name = r.FormValue("name")
	}

	if email == "" {
		http.Error(w, "Email required", 400)
		return
	}

	s.db.AddSubscriber(listID, email, name)

	accept := r.Header.Get("Accept")
	if strings.Contains(accept, "application/json") {
		writeJSON(w, 200, map[string]string{"status": "subscribed"})
		return
	}
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(`<!DOCTYPE html><html><head><title>Subscribed</title><style>body{font-family:system-ui;display:flex;justify-content:center;align-items:center;min-height:100vh;background:#1a1410;color:#f0e6d3}.box{text-align:center}h1{margin-bottom:.5rem}</style></head><body><div class="box"><h1>You're subscribed!</h1><p>Thanks for joining.</p></div></body></html>`))
}

func (s *Server) handleUnsubscribe(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "Missing token", 400)
		return
	}
	s.db.Unsubscribe(token)
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(`<!DOCTYPE html><html><head><title>Unsubscribed</title><style>body{font-family:system-ui;display:flex;justify-content:center;align-items:center;min-height:100vh;background:#1a1410;color:#f0e6d3}.box{text-align:center}h1{margin-bottom:.5rem}</style></head><body><div class="box"><h1>Unsubscribed</h1><p>You've been removed from the list.</p></div></body></html>`))
}

// --- Campaign handlers ---

func (s *Server) handleCreateCampaign(w http.ResponseWriter, r *http.Request) {
	listID := r.PathValue("id")
	if _, err := s.db.GetList(listID); err != nil {
		writeJSON(w, 404, map[string]string{"error": "list not found"})
		return
	}
	var req struct {
		Subject  string `json:"subject"`
		BodyHTML string `json:"body_html"`
		BodyText string `json:"body_text"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Subject == "" {
		writeJSON(w, 400, map[string]string{"error": "subject is required"})
		return
	}
	c, err := s.db.CreateCampaign(listID, req.Subject, req.BodyHTML, req.BodyText)
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, 201, map[string]any{"campaign": c})
}

func (s *Server) handleListCampaigns(w http.ResponseWriter, r *http.Request) {
	campaigns, err := s.db.ListCampaigns(r.PathValue("id"))
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	if campaigns == nil {
		campaigns = []store.Campaign{}
	}
	writeJSON(w, 200, map[string]any{"campaigns": campaigns, "count": len(campaigns)})
}

func (s *Server) handleGetCampaign(w http.ResponseWriter, r *http.Request) {
	c, err := s.db.GetCampaign(r.PathValue("id"))
	if err != nil {
		writeJSON(w, 404, map[string]string{"error": "campaign not found"})
		return
	}
	writeJSON(w, 200, map[string]any{"campaign": c})
}

func (s *Server) handleDeleteCampaign(w http.ResponseWriter, r *http.Request) {
	s.db.DeleteCampaign(r.PathValue("id"))
	writeJSON(w, 200, map[string]string{"status": "deleted"})
}

func (s *Server) handleListSends(w http.ResponseWriter, r *http.Request) {
	sends, err := s.db.ListSends(r.PathValue("id"))
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	if sends == nil {
		sends = []store.Send{}
	}
	writeJSON(w, 200, map[string]any{"sends": sends, "count": len(sends)})
}

// --- Send campaign ---

func (s *Server) handleSendCampaign(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	campaign, err := s.db.GetCampaign(id)
	if err != nil {
		writeJSON(w, 404, map[string]string{"error": "campaign not found"})
		return
	}
	if campaign.Status == "sent" {
		writeJSON(w, 400, map[string]string{"error": "campaign already sent"})
		return
	}

	if s.smtp.Host == "" {
		writeJSON(w, 400, map[string]string{"error": "SMTP not configured — set SMTP_HOST, SMTP_PORT, SMTP_USER, SMTP_PASS, SMTP_FROM"})
		return
	}

	subs, err := s.db.ActiveSubscribers(campaign.ListID)
	if err != nil || len(subs) == 0 {
		writeJSON(w, 400, map[string]string{"error": "no active subscribers"})
		return
	}

	s.db.UpdateCampaignStatus(id, "sending")

	go func() {
		for _, sub := range subs {
			send, _ := s.db.CreateSend(id, sub.ID, sub.Email)
			if send == nil {
				continue
			}

			// Build unsubscribe URL
			unsubURL := fmt.Sprintf("http://localhost:%d/unsubscribe?token=%s", s.port, sub.Token)

			body := campaign.BodyHTML
			if body == "" {
				body = campaign.BodyText
			}
			body += fmt.Sprintf(`<br><br><p style="font-size:12px;color:#999"><a href="%s">Unsubscribe</a></p>`, unsubURL)

			err := s.sendEmail(sub.Email, campaign.Subject, body)
			if err != nil {
				s.db.UpdateSendStatus(send.ID, "failed", err.Error())
				log.Printf("[send] %s → %s FAIL: %v", id, sub.Email, err)
			} else {
				s.db.UpdateSendStatus(send.ID, "sent", "")
				s.db.IncrementCampaignSent(id)
				log.Printf("[send] %s → %s OK", id, sub.Email)
			}

			time.Sleep(100 * time.Millisecond) // rate limit
		}
		s.db.UpdateCampaignStatus(id, "sent")
		log.Printf("[campaign] %s sent to %d subscribers", id, len(subs))
	}()

	writeJSON(w, 200, map[string]any{"status": "sending", "recipient_count": len(subs)})
}

func (s *Server) sendEmail(to, subject, bodyHTML string) error {
	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n%s",
		s.smtp.From, to, subject, bodyHTML)

	auth := smtp.PlainAuth("", s.smtp.User, s.smtp.Pass, s.smtp.Host)
	return smtp.SendMail(s.smtp.Host+":"+s.smtp.Port, auth, s.smtp.From, []string{to}, []byte(msg))
}

// --- Tracking ---

func (s *Server) handleTrackOpen(w http.ResponseWriter, r *http.Request) {
	campaignID := r.PathValue("id")
	s.db.RecordOpen(campaignID)
	// 1x1 transparent GIF
	w.Header().Set("Content-Type", "image/gif")
	w.Header().Set("Cache-Control", "no-store")
	w.Write([]byte{0x47, 0x49, 0x46, 0x38, 0x39, 0x61, 0x01, 0x00, 0x01, 0x00, 0x80, 0x00, 0x00, 0xff, 0xff, 0xff, 0x00, 0x00, 0x00, 0x21, 0xf9, 0x04, 0x01, 0x00, 0x00, 0x00, 0x00, 0x2c, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x01, 0x00, 0x00, 0x02, 0x02, 0x44, 0x01, 0x00, 0x3b})
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, s.db.Stats())
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, map[string]string{"status": "ok"})
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}
