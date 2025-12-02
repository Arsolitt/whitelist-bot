package kv

import (
	"context"
	"log/slog"
	"whitelist-bot/internal/core"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nkeys"
)

func GetNatsConn(ctx context.Context, cfg core.NatsConfig) (*nats.Conn, error) {
	sig, err := nkeys.FromSeed([]byte(cfg.NKeySeed))
	if err != nil {
		slog.ErrorContext(ctx, "Failed to create signature", "error", err)
	}
	return nats.Connect(cfg.URL, nats.Nkey(cfg.NKeyPublic, nats.SignatureHandler(func(nonce []byte) ([]byte, error) {
		sigData, err := sig.Sign(nonce)
		sig.Wipe()
		return sigData, err
	})))
}
