package sigstore

import (
	"github.com/gillisandrew/dragonglass-poc/internal/domain"
)

// Compile-time verification that Service implements domain.AttestationService
var _ domain.AttestationService = (*Service)(nil)