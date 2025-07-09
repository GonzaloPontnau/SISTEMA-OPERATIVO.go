package utils

import (
	"fmt"
	"os"
)

// IniciarServidor inicializa y arranca un servidor HTTP en una goroutine
func IniciarServidor(server *HTTPServer) {
	go func() {
		if err := server.Start(); err != nil {
			ErrorLog.Error("Error en servidor HTTP", "error", err)
			os.Exit(1)
		}
	}()
}

// Handler para handshake
func ManejarHandshake(msg *Mensaje) (interface{}, error) {
	InfoLog.Info("Handshake recibido",
		"origen", msg.Origen,
		"tipo", msg.Tipo,
		"operacion", msg.Operacion,
		"datos", fmt.Sprintf("%v", msg.Datos))
	return map[string]interface{}{
		"status": "OK",
		"modulo": "Servicio", // Valor genérico en lugar de usar Prefix que no existe
	}, nil
}
