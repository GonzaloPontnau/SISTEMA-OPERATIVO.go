package main

import (
	"fmt"

	"github.com/sisoputnfrba/tp-2025-1c-LosCuervosXeneizes/utils"
)

// Registrar todos los handlers de mensajes
func RegistrarHandlers() {
	modulo.RegistrarHandler(fmt.Sprintf("%d", utils.MensajeHandshake), "handshake", utils.ManejarHandshake)
	modulo.RegistrarHandler(fmt.Sprintf("%d", utils.MensajeOperacion), "EJECUTAR_PROCESO", manejarEjecutar)
	modulo.RegistrarHandler(fmt.Sprintf("%d", utils.MensajeEjecutar), "default", manejarEjecutar)
	modulo.RegistrarHandler(fmt.Sprintf("%d", utils.MensajeInterrupcion), "INTERRUPCION", manejarInterrupcion)
}

// Handler para ejecutar instrucción
func manejarEjecutar(msg *utils.Mensaje) (interface{}, error) {
	// Extraer PID y PC del mensaje
	datos := msg.Datos.(map[string]interface{})
	pid, okPid := datos["pid"].(float64)
	pc, okPc := datos["pc"].(float64)

	if !okPid || !okPc {
		utils.ErrorLog.Error("Formato de mensaje incorrecto", "datos", fmt.Sprintf("%v", datos))
		return map[string]interface{}{
			"error": "Formato de mensaje incorrecto",
		}, nil
	}

	pidInt := int(pid)
	pcInt := int(pc)

	utils.InfoLog.Info("Recibido proceso para ejecutar", "pid", pidInt, "pc", pcInt)

	// Ejecutar ciclo de instrucción
	siguientePC, motivo, parametrosSyscall := ejecutarCiclo(pidInt, pcInt)

	// Preparar respuesta
	respuesta := map[string]interface{}{
		"pid": pidInt,
		"pc":  siguientePC,
	}

	// Si hay un motivo de retorno (syscall, interrupción, error), agregarlo
	if motivo != "" {
		respuesta["motivo_retorno"] = motivo
		if parametrosSyscall != nil {
			respuesta["parametros"] = parametrosSyscall
		}
	}

	utils.InfoLog.Info("Proceso devuelto al Kernel", "pid", pidInt, "pc", siguientePC, "motivo", motivo)

	return respuesta, nil
}

// Handler para interrupciones
func manejarInterrupcion(msg *utils.Mensaje) (interface{}, error) {
	datos := msg.Datos.(map[string]interface{})
	pid, ok := datos["pid"].(float64)

	if !ok {
		utils.ErrorLog.Error("Formato de interrupción incorrecto", "datos", fmt.Sprintf("%v", datos))
		return map[string]interface{}{
			"error": "Formato de interrupción incorrecto",
		}, nil
	}

	pidInt := int(pid)

	mutex.Lock()
	interrupcionPendiente = true
	pidInterrumpido = pidInt
	mutex.Unlock()

	utils.InfoLog.Info("Interrupción recibida", "pid", pidInt)

	return map[string]interface{}{"ok": true}, nil
}
