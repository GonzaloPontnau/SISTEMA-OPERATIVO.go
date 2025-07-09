package main

import (
	"github.com/sisoputnfrba/tp-2025-1c-LosCuervosXeneizes/utils"
)

// Handlers
func handlerHandshake(msg *utils.Mensaje) (interface{}, error) {
	utils.InfoLog.Info("Handshake recibido", "origen", msg.Origen)
	return map[string]interface{}{"status": "OK"}, nil
}

func handlerOperacion(msg *utils.Mensaje) (interface{}, error) {
	return utils.HandlerGenerico(msg, config.RetardoBase, procesarOperacion)
}