package main

import (
	"fmt"
	"time"

	"github.com/sisoputnfrba/tp-2025-1c-LosCuervosXeneizes/utils"
)

var clienteKernel *utils.HTTPClient

// notificarIOTerminadaAKernel envía un mensaje al Kernel indicando que la operación IO ha terminado
func notificarIOTerminadaAKernel(pid int) {
	datos := map[string]interface{}{
		"evento":    "IO_TERMINADA",
		"operacion": "IO_COMPLETADA",
		"pid":       pid,
		"timestamp": time.Now().UnixNano() / int64(time.Millisecond),
	}

	// Enviar mensaje directamente usando la IP y puerto del Kernel desde la configuración
	cliente := modulo.Clientes["Kernel"]
	if cliente == nil {
		utils.ErrorLog.Error("No se pudo crear cliente para Kernel")
		return
	}

	_, err := cliente.EnviarHTTPOperacion("IO_COMPLETADA", datos)
	if err != nil {
		utils.ErrorLog.Error("Error notificando IO terminada a Kernel", "error", err.Error())
	} else {
		utils.InfoLog.Debug("Notificación IO terminada enviada a Kernel", "pid", pid)
	}
}

func procesarOperacion(msg *utils.Mensaje) (interface{}, error) {
	datos, ok := msg.Datos.(map[string]interface{})
	if !ok {
		utils.ErrorLog.Warn("Formato de datos inválido en operación IO")
		return map[string]interface{}{
			"status":  "ERROR",
			"mensaje": "Formato de datos inválido",
		}, nil
	}

	pidFloat, pidOk := datos["pid"].(float64)
	if !pidOk {
		utils.ErrorLog.Warn("Operación IO sin PID válido", "datos", datos)
		return map[string]interface{}{
			"status":  "ERROR",
			"mensaje": "PID inválido en solicitud IO",
		}, nil
	}
	pid := int(pidFloat)

	tiempoFloat, tiempoOk := datos["tiempo"].(float64)
	if !tiempoOk {
		utils.ErrorLog.Warn("Operación IO sin tiempo válido", "datos", datos)
		return map[string]interface{}{
			"status":  "ERROR",
			"mensaje": "Tiempo inválido en solicitud IO",
		}, nil
	}
	tiempo := int(tiempoFloat)

	// Log de inicio de IO con formato solicitado
	utils.InfoLog.Info(fmt.Sprintf("## PID: %d - Inicio de IO - Tiempo: %d", pid, tiempo))

	// Simular la operación IO con el retardo configurado
	utils.AplicarRetardo("io_operacion", tiempo)

	// Log de fin de IO con formato solicitado
	utils.InfoLog.Info(fmt.Sprintf("## PID: %d - Fin de IO", pid))

	// Notificar al Kernel que la operación IO ha terminado
	go notificarIOTerminadaAKernel(pid)

	return map[string]interface{}{
		"status":  "OK",
		"mensaje": "Operación I/O completada exitosamente",
	}, nil
}
