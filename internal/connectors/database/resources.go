package database

import (
	"context"
	"errors"
	"time"
)

const (
	hardMaximumReadPageSize     = 100_000
	hardMaximumWriteBatchSize   = 10_000
	hardMaximumConnectionPool   = 128
	hardMaximumBindParameters   = 65_535
	hardMaximumConnectTimeout   = 5 * time.Minute
	hardMaximumOperationTimeout = time.Hour
)

// Limit is a finite user-selectable resource limit. Default is applied only
// when a caller makes no selection; Maximum is always an enforced ceiling.
type Limit struct {
	Default int
	Maximum int
}

func (l Limit) validate(hardMaximum int) error {
	if l.Default <= 0 || l.Maximum <= 0 || l.Default > l.Maximum || l.Maximum > hardMaximum {
		return errors.New("database resource limit must have a positive finite default and maximum")
	}
	return nil
}

func (l Limit) resolve(requested int) (int, error) {
	if requested == 0 {
		return l.Default, nil
	}
	if requested < 0 || requested > l.Maximum {
		return 0, errors.New("database resource request exceeds the declared finite limit")
	}
	return requested, nil
}

// ResourcePolicy owns the non-HTTP resource bounds for a database driver. It
// contains no rate-limit placeholder: database connections must have concrete
// resource bounds rather than a guessed provider pacing policy.
type ResourcePolicy struct {
	ReadPage         Limit
	WriteBatch       Limit
	Pool             Limit
	ConnectTimeout   time.Duration
	OperationTimeout time.Duration
	MaxParameters    int
}

func (p ResourcePolicy) validate() error {
	if err := p.ReadPage.validate(hardMaximumReadPageSize); err != nil {
		return err
	}
	if err := p.WriteBatch.validate(hardMaximumWriteBatchSize); err != nil {
		return err
	}
	if err := p.Pool.validate(hardMaximumConnectionPool); err != nil {
		return err
	}
	if p.ConnectTimeout <= 0 || p.ConnectTimeout > hardMaximumConnectTimeout ||
		p.OperationTimeout <= 0 || p.OperationTimeout > hardMaximumOperationTimeout ||
		p.MaxParameters <= 0 || p.MaxParameters > hardMaximumBindParameters {
		return errors.New("database resource policy must use finite positive timeouts and parameter bounds")
	}
	return nil
}

// EffectivePageSize returns a bounded requested page size or the declaration's
// safe default when requested is zero.
func (p ResourcePolicy) EffectivePageSize(requested int) (int, error) {
	if err := p.validate(); err != nil {
		return 0, err
	}
	return p.ReadPage.resolve(requested)
}

// EffectiveBatchSize returns a bounded requested batch size or the safe
// declared default when requested is zero.
func (p ResourcePolicy) EffectiveBatchSize(requested int) (int, error) {
	if err := p.validate(); err != nil {
		return 0, err
	}
	return p.WriteBatch.resolve(requested)
}

// EffectivePoolSize returns a bounded requested pool size or the safe declared
// default when requested is zero.
func (p ResourcePolicy) EffectivePoolSize(requested int) (int, error) {
	if err := p.validate(); err != nil {
		return 0, err
	}
	return p.Pool.resolve(requested)
}

// WithOperationTimeout derives a bounded child context for a future driver
// operation. It does not open a connection or execute a query.
func (p ResourcePolicy) WithOperationTimeout(ctx context.Context) (context.Context, context.CancelFunc, error) {
	if ctx == nil {
		return nil, nil, errors.New("database operation context is required")
	}
	if err := ctx.Err(); err != nil {
		return nil, nil, err
	}
	if err := p.validate(); err != nil {
		return nil, nil, err
	}
	child, cancel := context.WithTimeout(ctx, p.OperationTimeout)
	return child, cancel, nil
}
