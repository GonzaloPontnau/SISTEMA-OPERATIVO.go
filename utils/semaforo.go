package utils

// Semaforo implementa un semáforo contador con canales
type Semaforo struct {
	c chan struct{}
}

// NewSemaforo crea un semáforo con capacidad inicial
func NewSemaforo(capacidad int) *Semaforo {
	if capacidad <= 0 {
		capacidad = 1
	}
	return &Semaforo{
		c: make(chan struct{}, capacidad),
	}
}

// Wait (P) decrementa el semáforo, bloquea si es 0
func (s *Semaforo) Wait() {
	s.c <- struct{}{} // Envía un valor, bloquea si está lleno
}

// Signal (V) incrementa el semáforo
func (s *Semaforo) Signal() {
	select {
	case <-s.c:
	default:
		// Capacidad completa, no hace nada para prevenir incremento excesivo
	}
}

// TryWait intenta decrementar sin bloquear
func (s *Semaforo) TryWait() bool {
	select {
	case s.c <- struct{}{}:
		return true
	default:
		return false
	}
}
