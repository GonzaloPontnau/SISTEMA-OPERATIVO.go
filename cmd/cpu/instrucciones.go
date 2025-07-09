package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/sisoputnfrba/tp-2025-1c-LosCuervosXeneizes/utils"
)

// Fetch: Obtener instrucción desde memoria
func fetch(pid, pc int) string {
	utils.InfoLog.Info(fmt.Sprintf("## PID: %d - FETCH - Program Counter: %d", pid, pc))

	// Preparar mensaje para memoria
	params := map[string]interface{}{
		"pid": pid,
		"pc":  pc,
	}

	// Enviar solicitud a memoria
	respuesta, err := modulo.EnviarMensaje("Memoria", utils.MensajeFetch, "FETCH", params)
	if err != nil {
		utils.ErrorLog.Error("Error al solicitar instrucción a memoria", "error", err)
		return ""
	}

	// Extraer instrucción de la respuesta
	respuestaMap, ok := respuesta.(map[string]interface{})
	if !ok {
		utils.ErrorLog.Error("Formato de respuesta incorrecto", "respuesta", fmt.Sprintf("%v", respuesta))
		return ""
	}

	instruccion, ok := respuestaMap["instruccion"].(string)
	if !ok {
		utils.ErrorLog.Error("Formato de instrucción incorrecto", "respuesta", fmt.Sprintf("%v", respuestaMap))
		return ""
	}

	utils.InfoLog.Info("Instrucción recibida de Memoria", "pid", pid, "pc", pc, "instruccion", instruccion)

	return instruccion
}

// Decode y Execute: Interpretar y ejecutar instrucción
func decodeAndExecute(pid, pc int, instruccion string) (int, string, map[string]interface{}) {
	partes := strings.Fields(instruccion)
	if len(partes) == 0 {
		utils.ErrorLog.Error("Instrucción vacía", "pid", pid, "pc", pc)
		return pc, "ERROR", nil
	}

	operacion := partes[0]
	parametros := partes[1:]

	utils.InfoLog.Info(fmt.Sprintf("## PID: %d - Ejecutando: %s - %s", pid, operacion, strings.Join(parametros, " ")))

	// Parámetros para syscall y motivo de retorno
	parametrosSyscall := make(map[string]interface{})
	motivoRetorno := ""
	siguientePC := pc

	// Ejecutar según el tipo de instrucción
	switch operacion {
	case "NOOP":
		// No hacer nada, sólo consumir tiempo

	case "WRITE":
		if len(parametros) >= 2 {
			direccion, err := strconv.Atoi(parametros[0])
			if err != nil {
				utils.ErrorLog.Error("Error en dirección WRITE", "error", err)
				motivoRetorno = "ERROR"
				break
			}
			datos := parametros[1]
			escribirEnMemoria(pid, direccion, datos)
		} else {
			utils.ErrorLog.Error("WRITE: parámetros insuficientes", "parametros", parametros)
			motivoRetorno = "ERROR"
		}

	case "READ":
		if len(parametros) >= 2 {
			direccion, err1 := strconv.Atoi(parametros[0])
			tamano, err2 := strconv.Atoi(parametros[1])
			if err1 != nil || err2 != nil {
				utils.ErrorLog.Error("Error en parámetros READ", "err1", err1, "err2", err2)
				motivoRetorno = "ERROR"
				break
			}
			leerDeMemoria(pid, direccion, tamano)
		} else {
			utils.ErrorLog.Error("READ: parámetros insuficientes", "parametros", parametros)
			motivoRetorno = "ERROR"
		}

	case "GOTO":
		if len(parametros) >= 1 {
			nuevoPC, err := strconv.Atoi(parametros[0])
			if err != nil {
				utils.ErrorLog.Error("Error en GOTO", "error", err)
				motivoRetorno = "ERROR"
				break
			}
			siguientePC = nuevoPC
		} else {
			utils.ErrorLog.Error("GOTO: parámetros insuficientes", "parametros", parametros)
			motivoRetorno = "ERROR"
		}

	case "IO":
		if len(parametros) >= 2 {
			dispositivo := parametros[0]
			tiempo, err := strconv.Atoi(parametros[1])
			if err != nil {
				utils.ErrorLog.Error("Error en tiempo IO", "error", err)
				motivoRetorno = "ERROR"
				break
			}
			parametrosSyscall["dispositivo"] = dispositivo
			parametrosSyscall["tiempo"] = tiempo
			motivoRetorno = "SYSCALL_IO"
		} else {
			utils.ErrorLog.Error("IO: parámetros insuficientes", "parametros", parametros)
			motivoRetorno = "ERROR"
		}

	case "INIT_PROC":
		if len(parametros) >= 2 {
			archivo := parametros[0]
			tamano, err := strconv.Atoi(parametros[1])
			if err != nil {
				utils.ErrorLog.Error("Error en tamaño INIT_PROC", "error", err)
				motivoRetorno = "ERROR"
				break
			}
			parametrosSyscall["archivo"] = archivo
			parametrosSyscall["tamano"] = tamano
			motivoRetorno = "SYSCALL_INIT_PROC"
		} else {
			utils.ErrorLog.Error("INIT_PROC: parámetros insuficientes", "parametros", parametros)
			motivoRetorno = "ERROR"
		}

	case "DUMP_MEMORY":
		motivoRetorno = "SYSCALL_DUMP_MEMORY"

	case "EXIT":
		motivoRetorno = "EXIT"

	default:
		utils.ErrorLog.Error("Instrucción desconocida", "operacion", operacion)
		motivoRetorno = "ERROR"
	}

	return siguientePC, motivoRetorno, parametrosSyscall
}
