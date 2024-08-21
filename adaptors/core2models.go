package adaptors

import (
	"github.com/NethermindEth/juno/core"
)

type PostgresAdapter struct {
}

func (p *PostgresAdapter) onDeposit(event core.Event) error {

	return nil
}
