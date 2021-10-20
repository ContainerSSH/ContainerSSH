package none

import (
	"context"
)

type nopStorage struct {
}

func (s *nopStorage) Shutdown(_ context.Context) {

}
