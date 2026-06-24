package store

import (
	"testing"
	"time"

	"restdeck/internal/domain"
)

func TestStoreCookies(t *testing.T) {
	ctx := t.Context()
	s, err := OpenInMemory(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	cookie := domain.Cookie{
		Name:     "sid",
		Value:    "abc",
		Domain:   "example.com",
		Path:     "/",
		Expires:  time.Now().Add(time.Hour),
		HTTPOnly: true,
		Secure:   true,
	}
	if err := s.SaveCookies(ctx, []domain.Cookie{cookie}); err != nil {
		t.Fatal(err)
	}
	cookies, err := s.ListCookies(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(cookies) != 1 || cookies[0].Name != "sid" || cookies[0].Value != "abc" {
		t.Fatalf("unexpected cookies: %#v", cookies)
	}

	if err := s.DeleteCookie(ctx, cookie); err != nil {
		t.Fatal(err)
	}
	cookies, err = s.ListCookies(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(cookies) != 0 {
		t.Fatalf("cookie not deleted: %#v", cookies)
	}

	if err := s.SaveCookies(ctx, []domain.Cookie{cookie}); err != nil {
		t.Fatal(err)
	}
	if err := s.ClearCookies(ctx); err != nil {
		t.Fatal(err)
	}
	cookies, err = s.ListCookies(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(cookies) != 0 {
		t.Fatalf("cookies not cleared: %#v", cookies)
	}
}
