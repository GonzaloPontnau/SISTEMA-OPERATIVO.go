package main

import (
	"fmt"
	"strconv"

	"github.com/sisoputnfrba/tp-2025-1c-LosCuervosXeneizes/utils"
)

// HandlerHandshake simplificado y optimizado
func HandlerHandshake(msg *utils.Mensaje) (interface{}, error) {
	// DEBUG: Log de todo lo que llega
	utils.InfoLog.Info("HandlerHandshake recibido", "origen", msg.Origen, "datos", fmt.Sprintf("%v", msg.Datos))

	datosMap, ok := msg.Datos.(map[string]interface{})
	if !ok {
		utils.ErrorLog.Error("HandlerHandshake: Datos inválidos", "datos", fmt.Sprintf("%v", msg.Datos))
		return map[string]interface{}{"status": "ERROR", "message": "Datos inválidos"}, nil
	}

	// Procesar en paralelo cuando sea posible
	resultChan := make(chan interface{}, 1)

	go func() {
		// Verificar IO primero
		if respuesta, manejado := ManejadorRegistroIO(msg.Origen, datosMap); manejado {
			utils.InfoLog.Info("HandlerHandshake: Procesado como IO", "origen", msg.Origen)
			resultChan <- respuesta
			return
		}

		// Verificar CPU
		esCPUResult := esCPU(msg.Origen, datosMap)
		utils.InfoLog.Info("HandlerHandshake: Verificando si es CPU", "origen", msg.Origen, "esCPU", esCPUResult)

		if esCPUResult {
			utils.InfoLog.Info("HandlerHandshake: Procesando como CPU", "origen", msg.Origen)
			respuesta, _ := manejarRegistroCPU(msg.Origen, datosMap)
			resultChan <- respuesta
			return
		}

		utils.InfoLog.Info("HandlerHandshake: Handshake genérico", "origen", msg.Origen)
		resultChan <- map[string]interface{}{"status": "OK", "message": "Handshake recibido"}
	}()

	return <-resultChan, nil
}

// esCPU simplificado con evaluación corta
func esCPU(origen string, datos map[string]interface{}) bool {
	return origen == "CPU" ||
		datos["tipo"] == "CPU" ||
		datos["nombre"] == "CPU"
}

// manejarRegistroCPU optimizado con mejor manejo de tipos
func manejarRegistroCPU(origen string, datos map[string]interface{}) (interface{}, error) {
	ip, ipOk := datos["ip"].(string)
	if !ipOk {
		return map[string]interface{}{"status": "ERROR", "message": "IP requerida"}, nil
	}

	puerto, puertoOk := extraerPuerto(datos["puerto"])
	if !puertoOk {
		return map[string]interface{}{"status": "ERROR", "message": "Puerto inválido"}, nil
	}

	// Usar el identificador específico de la CPU en lugar del origen genérico
	identificadorCPU := origen // Por defecto, usar el origen
	if id, existe := datos["identificador"].(string); existe && id != "" {
		identificadorCPU = id
	}

	// El registro DEBE ser síncrono para evitar race conditions con el planificador.
	registrarCPU(identificadorCPU, ip, puerto)

	// Log original mantenido exactamente igual
	utils.InfoLog.Info(fmt.Sprintf("CPU %s registrada", identificadorCPU), "ip", ip, "puerto", puerto)

	return map[string]interface{}{"status": "OK", "message": fmt.Sprintf("CPU %s registrada", identificadorCPU)}, nil
}

// extraerPuerto helper para simplificar conversión de tipos
func extraerPuerto(puerto interface{}) (int, bool) {
	switch p := puerto.(type) {
	case float64:
		return int(p), true
	case int:
		return p, true
	case string:
		if val, err := strconv.Atoi(p); err == nil {
			return val, true
		}
	}
	return 0, false
}

func HandlerOperacion(msg *utils.Mensaje) (interface{}, error) {
	return procesarOperacionEspecifica(msg)
}

// procesarOperacionEspecifica optimizado con pipeline de procesamiento
func procesarOperacionEspecifica(msg *utils.Mensaje) (interface{}, error) {
	datos, ok := msg.Datos.(map[string]interface{})
	if !ok {
		return map[string]interface{}{"status": "ERROR", "mensaje": "Datos inválidos"}, nil
	}

	pid, pidOk := extraerPID(datos["pid"])
	if !pidOk {
		// Si no hay PID, podría ser una operación que no lo requiere
		// o un error. Por ahora, asumimos que la mayoría lo necesita.
		// NOTA: El handshake de IO no pasa por aquí.
		if _, esFinIO := datos["evento"].(string); !esFinIO {
			return map[string]interface{}{"status": "ERROR", "mensaje": "PID inválido o faltante"}, nil
		}
	}

	// Pipeline de procesamiento optimizado
	handlers := []func(int, map[string]interface{}) (interface{}, bool){
		ProcesarRetornoCPU,
		func(pid int, datos map[string]interface{}) (interface{}, bool) {
			return ProcesarSolicitudIO(datos)
		},
		func(pid int, datos map[string]interface{}) (interface{}, bool) {
			return ProcesarIOTerminada(datos)
		},
		procesarFinalizacionSiCorresponde,
	}

	for _, handler := range handlers {
		if respuesta, manejado := handler(pid, datos); manejado {
			return respuesta, nil
		}
	}

	return map[string]interface{}{"status": "ERROR", "mensaje": "Operación desconocida o no manejada"}, nil
}

// ProcesarRetornoCPU maneja el retorno de un proceso desde la CPU que no sea finalización.
func ProcesarRetornoCPU(pid int, datos map[string]interface{}) (interface{}, bool) {
	motivo, ok := datos["motivo_retorno"].(string)
	if !ok {
		return nil, false // No es un retorno de CPU o no tiene motivo
	}

	pcb := BuscarPCBPorPID(pid)
	if pcb == nil {
		utils.ErrorLog.Warn("Se recibió retorno de CPU para un PID inexistente", "pid", pid)
		return map[string]interface{}{"status": "ERROR", "mensaje": "PID no encontrado"}, true
	}

	// Liberar la CPU
	liberarCPU(pid)

	switch motivo {
	case "INTERRUPTED":
		utils.InfoLog.Info(fmt.Sprintf("## (%d) - Proceso interrumpido por Kernel", pid))
		MoverProcesoAReady(pcb)
		go despacharProcesoSiCorresponde() // Intentar despachar otro proceso si hay CPU libre
		return map[string]interface{}{"status": "OK", "message": "Proceso movido a READY por interrupción"}, true

	case "SYSCALL_IO":
		// La lógica de IO ya está en ProcesarSolicitudIO.
		// Aquí solo actuamos si el evento es específico de retorno de CPU.
		// Podríamos unificarlo, pero por ahora lo dejamos separado.
		// Este caso es manejado por ProcesarSolicitudIO.
		return nil, false

	default:
		// Otros motivos como EXIT o ERROR son manejados por procesarFinalizacionSiCorresponde
		return nil, false
	}
}

// liberarCPU encuentra y libera la CPU que estaba ejecutando un proceso.
func liberarCPU(pid int) string {
	execMutex.Lock()
	defer execMutex.Unlock()

	var cpuLiberada string
	for cpu, pcbEnExec := range colaExec {
		if pcbEnExec != nil && pcbEnExec.PID == pid {
			delete(colaExec, cpu)
			cpuLiberada = cpu
			break
		}
	}
	if cpuLiberada != "" {
		utils.InfoLog.Debug("CPU liberada", "cpu", cpuLiberada, "pid", pid)
	}
	return cpuLiberada
}

// extraerPID helper simplificado
func extraerPID(pid interface{}) (int, bool) {
	if pidFloat, ok := pid.(float64); ok {
		return int(pidFloat), true
	}
	return 0, false
}

// actualizarPCDesdeCPU función asíncrona para actualización de PC
func actualizarPCDesdeCPU(pid int, datos map[string]interface{}) {
	if pcb := BuscarPCBPorPID(pid); pcb != nil {
		if pcActualizado, ok := datos["pc_actualizado"].(float64); ok {
			pcb.PC = int(pcActualizado)
		}
	}
}

// procesarFinalizacionSiCorresponde simplificado con detección mejorada
func procesarFinalizacionSiCorresponde(pid int, datos map[string]interface{}) (interface{}, bool) {
	evento, _ := datos["evento"].(string)
	motivoRetorno, _ := datos["motivo_retorno"].(string)

	if evento == "PROCESO_TERMINADO" || motivoRetorno == "EXIT" || motivoRetorno == "ERROR" {
		// Liberar la CPU antes de finalizar
		liberarCPU(pid)
		respuesta, _ := procesarFinalizacion(pid, datos)
		return respuesta, true
	}
	return nil, false
}

// procesarFinalizacion optimizado con operaciones paralelas
func procesarFinalizacion(pid int, datos map[string]interface{}) (interface{}, error) {
	pcb := BuscarPCBPorPID(pid)
	if pcb == nil {
		return map[string]interface{}{"status": "ERROR", "mensaje": "Proceso no encontrado"}, nil
	}

	motivo := determinarMotivo(datos)

	FinalizarProceso(pcb, motivo)

	// Operaciones post-finalización en paralelo
	go func() {
		intentarAdmitirProceso()
		despacharProcesoSiCorresponde()
	}()

	return map[string]interface{}{"status": "OK", "mensaje": "Proceso finalizado"}, nil
}

// determinarMotivo helper para lógica de motivo
func determinarMotivo(datos map[string]interface{}) string {
	if m, ok := datos["motivo"].(string); ok {
		return m
	}
	if motivoRetorno, ok := datos["motivo_retorno"].(string); ok && motivoRetorno == "ERROR" {
		return "ERROR_CPU"
	}
	return "EXIT_NORMAL"
}