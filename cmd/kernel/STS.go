package main

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/sisoputnfrba/tp-2025-1c-LosCuervosXeneizes/utils"
)

// Variables globales para gestión de CPUs
var (
	cpuClients               map[string]*utils.HTTPClient
	cpuClientsMutex          sync.Mutex
	ultimoLogCPUNoDisponible time.Time // Para evitar spam de logs
)

// InicializarMapaCPUs inicializa el mapa de CPUs de forma segura
func InicializarMapaCPUs() {
	cpuClientsMutex.Lock()
	defer cpuClientsMutex.Unlock()

	if cpuClients == nil {
		cpuClients = make(map[string]*utils.HTTPClient)
		utils.InfoLog.Info("Mapa cpuClients inicializado durante el arranque del kernel")
	}
}

// registrarCPU registra un nuevo cliente CPU
func registrarCPU(nombre string, ip string, puerto int) {
	cpuClientsMutex.Lock()
	defer cpuClientsMutex.Unlock()

	// DIAGNÓSTICO: Verificar estado inicial del mapa
	utils.InfoLog.Debug("Estado del mapa cpuClients ANTES del registro", "total_cpus", len(cpuClients), "keys", obtenerNombresCPUs())

	// El mapa ya debe estar inicializado durante el arranque del kernel
	if cpuClients == nil {
		utils.ErrorLog.Error("ERROR CRÍTICO: Mapa cpuClients no inicializado - esto indica un problema en el arranque del kernel")
		cpuClients = make(map[string]*utils.HTTPClient) // Fallback de emergencia
	}

	// Verificar si la CPU ya estaba registrada
	_, existe := cpuClients[nombre]
	if existe {
		utils.InfoLog.Info(fmt.Sprintf("CPU '%s' ya registrada, actualizando conexión", nombre))
	}

	// Registrar la CPU - Usar una clave más distintiva para evitar colisiones
	nombreCPU := nombre
	if nombreCPU == "" {
		nombreCPU = fmt.Sprintf("CPU_%s_%d", ip, puerto)
		utils.InfoLog.Info("Usando nombre generado para CPU", "nombre_generado", nombreCPU)
	}

	// Crear el cliente HTTP para esta CPU
	cpuClients[nombreCPU] = utils.NewHTTPClient(ip, puerto, "Kernel->"+nombreCPU)

	utils.InfoLog.Info(fmt.Sprintf("## CPU '%s' registrada correctamente en %s:%d", nombreCPU, ip, puerto),
		"total_cpus", len(cpuClients))

	// Mostrar todas las CPUs registradas actualmente
	utils.InfoLog.Debug("CPUs registradas DESPUÉS del registro:", "total", len(cpuClients), "nombres", obtenerNombresCPUs())
}

// obtenerNombresCPUs devuelve una lista con los nombres de todas las CPUs registradas
func obtenerNombresCPUs() []string {
	nombres := make([]string, 0, len(cpuClients))
	for nombre := range cpuClients {
		nombres = append(nombres, nombre)
	}
	return nombres
}

// obtenerCPUDisponible retorna un cliente de CPU disponible o nil si no hay ninguno
func obtenerCPUDisponible() (string, *utils.HTTPClient) {
	cpuClientsMutex.Lock()
	defer cpuClientsMutex.Unlock()

	// DIAGNÓSTICO: Información detallada sobre el estado del mapa
	cpuNombres := obtenerNombresCPUs()
	utils.InfoLog.Debug("Verificando CPUs disponibles", "total_registradas", len(cpuClients), "nombres", cpuNombres)

	// Verificar si el mapa está inicializado
	if cpuClients == nil {
		utils.InfoLog.Error("Mapa cpuClients no inicializado, esto no debería ocurrir")
		return "", nil
	}

	if len(cpuClients) == 0 {
		// Limitar los mensajes de log para evitar spam (máximo uno cada 5 segundos)
		ahora := time.Now()
		if ahora.Sub(ultimoLogCPUNoDisponible) > 5*time.Second {
			utils.InfoLog.Warn("No hay CPUs registradas en el sistema")
			ultimoLogCPUNoDisponible = ahora
		}
		return "", nil
	}

	// Por simplicidad, retornamos la primera CPU (en un sistema real, se implementaría un algoritmo de selección)
	for nombre, cliente := range cpuClients {
		utils.InfoLog.Info("Seleccionando CPU para ejecución", "nombre", nombre)
		return nombre, cliente
	}

	// Si por alguna razón no se encontró ninguna (lo cual no debería ocurrir si len > 0)
	utils.InfoLog.Error("Error lógico: len(cpuClients) > 0 pero no se encontró ninguna CPU")
	return "", nil
}

// PlanificarCortoPlazo gestiona transición de procesos entre READY y EXEC
func PlanificarCortoPlazo() {
	// Recuperación de pánico para diagnóstico
	defer func() {
		if r := recover(); r != nil {
			utils.ErrorLog.Error("PÁNICO EN PLANIFICADOR DE CORTO PLAZO", "error", r)
			// Re-panic para terminar el programa y mostrar el stack trace
			panic(r)
		}
	}()

	utils.InfoLog.Info("STS: Iniciando Planificador de Corto Plazo...")

	for {
		utils.InfoLog.Debug("STS: Esperando procesos en READY...")
		readyMutex.Lock()
		for len(colaReady) == 0 {
			condReady.Wait()
		}
		utils.InfoLog.Info("STS: Proceso detectado en READY, procediendo a seleccionar...", "procesos_en_ready", len(colaReady))

		// Seleccionar próximo proceso según algoritmo
		utils.InfoLog.Info("STS: Seleccionando proceso según algoritmo")
		pcb := seleccionarProcesoSTS()

		if pcb != nil {
			utils.InfoLog.Info("STS: Proceso seleccionado", "pcb_pid", pcb.PID)
			// Remover inmediatamente de READY para evitar condiciones de carrera
			// Ya tenemos el readyMutex bloqueado, usar directamente removerDeCola
			removerDeCola(&colaReady, pcb)
		} else {
			utils.InfoLog.Info("STS: Proceso seleccionado es nil")
		}

		// Si SRT decide desalojar, pcb será nil. El planificador se re-ejecutará
		if pcb == nil {
			utils.InfoLog.Debug("STS: Proceso seleccionado es nil, continuando")
			readyMutex.Unlock()
			time.Sleep(100 * time.Millisecond) // Pequeña pausa para que el desalojo se procese
			continue
		}

		readyMutex.Unlock() // Desbloquear la cola de ready una vez tenemos un candidato

		// Bucle para esperar una CPU disponible
		var nombreCPU string
		var cpuClient *utils.HTTPClient

		utils.InfoLog.Info("STS: Buscando CPU disponible...")
		for {
			nombreCPU, cpuClient = obtenerCPUDisponibleParaEjecucion()
			utils.InfoLog.Info("STS: Resultado búsqueda CPU", "nombre", nombreCPU, "cliente_encontrado", cpuClient != nil)
			if cpuClient != nil {
				// CPU encontrada. Reservarla.
				execMutex.Lock()
				colaExec[nombreCPU] = pcb
				execMutex.Unlock()
				utils.InfoLog.Info("STS: CPU encontrada y reservada", "nombre", nombreCPU)
				break
			}
			utils.InfoLog.Warn("STS: No hay CPU disponible, reintentando en 200ms...")
			time.Sleep(200 * time.Millisecond) // Esperar antes de volver a verificar
		}

		// Asignar a CPU
		utils.InfoLog.Info("STS: Cambiando estado del proceso a EXEC", "pid", pcb.PID)
		pcb.CambiarEstado(EstadoExec)
		utils.InfoLog.Info("STS: Estado del proceso cambiado a EXEC", "pid", pcb.PID)

		utils.InfoLog.Info(fmt.Sprintf("Proceso despachado a CPU %s", nombreCPU), "pid", pcb.PID)

		// Enviar a CPU y procesar su ciclo de ejecución de forma síncrona en una goroutine
		utils.InfoLog.Info("STS: Iniciando goroutine despacharYProcesarCPU", "pid", pcb.PID, "cpu", nombreCPU)
		go despacharYProcesarCPU(nombreCPU, cpuClient, pcb)
	}
}

// despacharYProcesarCPU se encarga del ciclo de vida de un proceso en la CPU
func despacharYProcesarCPU(nombreCPU string, cpuClient *utils.HTTPClient, pcb *PCB) {
	utils.InfoLog.Info("despacharYProcesarCPU: Iniciando", "pid", pcb.PID, "cpu", nombreCPU)

	// Liberar la CPU al final de la ejecución, sin importar cómo termine
	defer func() {
		utils.InfoLog.Info("despacharYProcesarCPU: Liberando CPU en defer", "pid", pcb.PID, "cpu", nombreCPU)
		execMutex.Lock()
		delete(colaExec, nombreCPU)
		execMutex.Unlock()
		utils.InfoLog.Info(fmt.Sprintf("CPU %s liberada", nombreCPU), "pid", pcb.PID)
	}()

	// Ciclo de ejecución en CPU - continúa mientras el proceso esté en EXEC
	for pcb.Estado == EstadoExec {
		utils.InfoLog.Info("despacharYProcesarCPU: Enviando proceso a CPU", "pid", pcb.PID, "cpu", nombreCPU, "pc", pcb.PC)
		fueExitoso := EnviarProcesoCPU(pcb, nombreCPU)
		utils.InfoLog.Info("despacharYProcesarCPU: Resultado de EnviarProcesoCPU", "pid", pcb.PID, "exitoso", fueExitoso, "estado_actual", pcb.Estado)

		// Si la comunicación falló o la CPU reportó un error, salir del ciclo
		if !fueExitoso {
			utils.ErrorLog.Error("El ciclo de ejecución en CPU falló", "pid", pcb.PID, "cpu", nombreCPU)
			// Si la comunicación falló, EnviarProcesoCPU ya se encargó del estado del PCB
			break
		}

		// Si el proceso cambió de estado (por syscall, exit, etc.), salir del ciclo
		if pcb.Estado != EstadoExec {
			utils.InfoLog.Info("despacharYProcesarCPU: Proceso cambió de estado, finalizando ciclo", "pid", pcb.PID, "nuevo_estado", pcb.Estado)
			break
		}
	}
	// Si fue exitoso, el proceso ya fue transicionado a otro estado (Blocked, Exit, etc)
	// por la función EnviarProcesoCPU.
}

// obtenerCPUDisponibleParaEjecucion busca una CPU que no esté en la cola de ejecución.
// Esta función ahora es segura para ser llamada sin un bloqueo externo en execMutex.
func obtenerCPUDisponibleParaEjecucion() (string, *utils.HTTPClient) {
	utils.InfoLog.Info("STS: Iniciando búsqueda de CPU disponible")

	// 1. Obtener una lista de todas las CPUs registradas (lectura segura)
	cpuClientsMutex.Lock()
	cpusDisponibles := make(map[string]*utils.HTTPClient)
	for nombre, cliente := range cpuClients {
		cpusDisponibles[nombre] = cliente
	}
	cpuClientsMutex.Unlock()

	utils.InfoLog.Info("STS: CPUs registradas obtenidas", "total", len(cpusDisponibles))

	if len(cpusDisponibles) == 0 {
		if time.Since(ultimoLogCPUNoDisponible) > 5*time.Second {
			utils.InfoLog.Warn("No hay CPUs registradas en el sistema.")
			ultimoLogCPUNoDisponible = time.Now()
		}
		return "", nil
	}

	// 2. Verificar cuáles de ellas no están en la cola de ejecución (lectura segura)
	execMutex.Lock()
	utils.InfoLog.Info("STS: Verificando CPUs en ejecución", "total_en_exec", len(colaExec))
	for cpuNombre, pcb := range colaExec {
		if pcb != nil {
			utils.InfoLog.Info("STS: CPU ocupada", "nombre", cpuNombre, "pid", pcb.PID)
		} else {
			utils.InfoLog.Info("STS: CPU en colaExec pero con PCB nil", "nombre", cpuNombre)
		}
	}
	execMutex.Unlock()

	execMutex.Lock()
	defer execMutex.Unlock()

	for nombre, cliente := range cpusDisponibles {
		if _, ocupada := colaExec[nombre]; !ocupada {
			utils.InfoLog.Info("STS: Se encontró CPU libre", "nombre", nombre)
			return nombre, cliente // Encontramos una CPU libre
		} else {
			utils.InfoLog.Info("STS: CPU está ocupada", "nombre", nombre)
		}
	}

	// 3. Si llegamos aquí, todas las CPUs registradas están ocupadas.
	utils.InfoLog.Info("STS: Todas las CPUs están ocupadas")
	return "", nil
}

// seleccionarProcesoSTS selecciona el próximo proceso de cola READY según algoritmo configurado
// ¡IMPORTANTE! Esta función debe ser llamada con el mutex de colaReady BLOQUEADO.
func seleccionarProcesoSTS() *PCB {
	if len(colaReady) == 0 {
		return nil
	}

	algoritmo := kernelConfig.SchedulerAlgorithm
	utils.InfoLog.Debug("STS seleccionando proceso", "algoritmo", algoritmo, "procesos_disponibles", len(colaReady))

	switch algoritmo {
	case "FIFO":
		return seleccionarFIFO()
	case "SJF":
		return seleccionarSJF()
	case "SRT":
		return seleccionarSRT()
	default:
		utils.InfoLog.Warn("Algoritmo STS no reconocido, usando FIFO", "algoritmo", algoritmo)
		return seleccionarFIFO()
	}
}

// seleccionarFIFO implementa selección FIFO
func seleccionarFIFO() *PCB {
	if len(colaReady) == 0 {
		return nil
	}
	return colaReady[0]
}

// seleccionarSJF implementa Shortest Job First sin desalojo
func seleccionarSJF() *PCB {
	if len(colaReady) == 0 {
		return nil
	}

	// Crear copia para ordenar
	candidatos := make([]*PCB, len(colaReady))
	copy(candidatos, colaReady)

	// Ordenar por estimación de tiempo (menor = mayor prioridad)
	sort.Slice(candidatos, func(i, j int) bool {
		// Si tienen la misma estimación, usar FIFO
		if candidatos[i].EstimacionSiguienteRafaga == candidatos[j].EstimacionSiguienteRafaga {
			return candidatos[i].HoraListo.Before(candidatos[j].HoraListo)
		}
		return candidatos[i].EstimacionSiguienteRafaga < candidatos[j].EstimacionSiguienteRafaga
	})

	seleccionado := candidatos[0]
	utils.InfoLog.Debug("SJF seleccionó proceso",
		"pid", seleccionado.PID,
		"estimacion", seleccionado.EstimacionSiguienteRafaga)

	return seleccionado
}

// seleccionarSRT implementa Shortest Remaining Time (SJF con desalojo)
// ¡IMPORTANTE! Esta función debe ser llamada con el mutex de colaReady BLOQUEADO.
func seleccionarSRT() *PCB {
	if len(colaReady) == 0 {
		return nil
	}

	// Encontrar el proceso con menor tiempo estimado en READY
	mejorCandidatoReady := encontrarMejorCandidatoReady()

	// Comparar con los procesos en ejecución
	execMutex.Lock()
	procesoADesalojar := encontrarProcesoADesalojar(mejorCandidatoReady)
	execMutex.Unlock()

	if procesoADesalojar != nil {
		utils.InfoLog.Info(fmt.Sprintf("## (%d) - Desalojado por algoritmo SJF/SRT por proceso (%d)", procesoADesalojar.PID, mejorCandidatoReady.PID))
		go desalojarProcesoActual(procesoADesalojar)
		return nil
	}

	return mejorCandidatoReady
}

// ¡IMPORTANTE! Esta función debe ser llamada con el mutex de colaReady BLOQUEADO.
func encontrarMejorCandidatoReady() *PCB {
	if len(colaReady) == 0 {
		return nil
	}
	// Encontrar el proceso con menor tiempo estimado en READY
	mejorProceso := colaReady[0]
	for _, pcb := range colaReady[1:] {
		if pcb.EstimacionSiguienteRafaga < mejorProceso.EstimacionSiguienteRafaga {
			mejorProceso = pcb
		} else if pcb.EstimacionSiguienteRafaga == mejorProceso.EstimacionSiguienteRafaga {
			if pcb.HoraListo.Before(mejorProceso.HoraListo) {
				mejorProceso = pcb
			}
		}
	}
	return mejorProceso
}

// ¡IMPORTANTE! Esta función debe ser llamada con el mutex de execMutex BLOQUEADO.
func encontrarProcesoADesalojar(candidato *PCB) *PCB {
	var procesoMasLargo *PCB = nil
	if candidato == nil {
		return nil
	}

	for _, pcbEnExec := range colaExec {
		if candidato.EstimacionSiguienteRafaga < pcbEnExec.EstimacionSiguienteRafaga {
			if procesoMasLargo == nil || pcbEnExec.EstimacionSiguienteRafaga > procesoMasLargo.EstimacionSiguienteRafaga {
				procesoMasLargo = pcbEnExec
			}
		}
	}
	return procesoMasLargo
}

// desalojarProcesoActual maneja el desalojo de un proceso por SRT
func desalojarProcesoActual(pcb *PCB) {
	var cpuADesalojar string
	execMutex.Lock()
	// Verificar que sigue siendo el proceso en ejecución y encontrar su CPU
	for cpu, pcbEnExec := range colaExec {
		if pcbEnExec != nil && pcbEnExec.PID == pcb.PID {
			cpuADesalojar = cpu
			break
		}
	}
	execMutex.Unlock()

	if cpuADesalojar == "" {
		utils.InfoLog.Warn("SRT: Intento de desalojar proceso que ya no está en ejecución", "pid", pcb.PID)
		return
	}

	cpuClientsMutex.Lock()
	cpuClient, existe := cpuClients[cpuADesalojar]
	cpuClientsMutex.Unlock()

	if !existe {
		utils.ErrorLog.Error("SRT: No se encontró el cliente para la CPU a desalojar", "cpu", cpuADesalojar)
		return
	}

	// Enviar interrupción a la CPU
	utils.InfoLog.Info("SRT: Enviando interrupción a CPU", "cpu", cpuADesalojar, "pid", pcb.PID)
	_, err := cpuClient.EnviarHTTPOperacion("INTERRUPT", nil) // No necesita payload
	if err != nil {
		utils.ErrorLog.Error("SRT: Fallo al enviar interrupción a la CPU", "cpu", cpuADesalojar, "error", err)
		// Aquí podría haber una lógica de reintento o manejo de fallo de CPU
	}
}

// EnviarProcesoCPU envía un PCB a la CPU para su ejecución
func EnviarProcesoCPU(pcb *PCB, nombreCPU string) bool {
	utils.InfoLog.Info("EnviarProcesoCPU: Iniciando envío", "pid", pcb.PID, "cpu", nombreCPU)

	// Obtener cliente para esta CPU
	cpuClientsMutex.Lock()
	cpuClient, existe := cpuClients[nombreCPU]
	cpuClientsMutex.Unlock()

	utils.InfoLog.Info("EnviarProcesoCPU: Cliente CPU obtenido", "pid", pcb.PID, "cpu", nombreCPU, "existe", existe)

	if !existe {
		utils.ErrorLog.Error("CPU no encontrada", "cpu_id", nombreCPU)
		return false
	}

	// Preparar datos para enviar a la CPU
	datos := map[string]interface{}{
		"pid": pcb.PID,
		"pc":  pcb.PC,
	}

	// Log del envío
	utils.InfoLog.Info(fmt.Sprintf("## (%d) - Enviando proceso a CPU", pcb.PID),
		"pc", pcb.PC, "cpu", nombreCPU)

	utils.InfoLog.Info("EnviarProcesoCPU: Realizando solicitud HTTP", "pid", pcb.PID, "cpu", nombreCPU)

	// Enviar solicitud a la CPU
	respuesta, err := cpuClient.EnviarHTTPOperacion("EJECUTAR_PROCESO", datos)

	utils.InfoLog.Info("EnviarProcesoCPU: Solicitud HTTP completada", "pid", pcb.PID, "cpu", nombreCPU, "error", err != nil)

	if err != nil {
		utils.ErrorLog.Error("Error enviando proceso a CPU",
			"pid", pcb.PID,
			"error", err.Error())

		// Devolver proceso a READY si falla la comunicación
		MoverProcesoAReady(pcb)

		return false
	}

	// Procesar respuesta
	if respuestaMap, ok := respuesta.(map[string]interface{}); ok {
		// Verificar si hay un error específico
		if errorMsg, tieneError := respuestaMap["error"].(string); tieneError {
			utils.ErrorLog.Error("Error reportado por CPU",
				"pid", pcb.PID,
				"mensaje", errorMsg)
			return false
		}

		// Actualizaciones según la respuesta
		utils.InfoLog.Debug("Respuesta de CPU recibida", "respuesta", fmt.Sprintf("%+v", respuestaMap))

		// Actualizar PC y marcar si fue actualizado por CPU
		pcActualizadoPorCPU := false
		if nuevoPC, hayPC := respuestaMap["pc"].(float64); hayPC {
			pcb.PC = int(nuevoPC)
			pcActualizadoPorCPU = true
			utils.InfoLog.Debug("PC actualizado desde respuesta CPU", "pid", pcb.PID, "nuevo_pc", pcb.PC)
		}

		// Verificar motivo de retorno para syscalls
		if motivoRetorno, hayMotivo := respuestaMap["motivo_retorno"].(string); hayMotivo {
			utils.InfoLog.Info(fmt.Sprintf("## (%d) - Motivo de retorno recibido", pcb.PID), "motivo", motivoRetorno)

			switch motivoRetorno {
			case "SYSCALL_INIT_PROC":
				// Procesar creación de nuevo proceso
				if parametros, ok := respuestaMap["parametros"].(map[string]interface{}); ok {
					archivo, _ := parametros["archivo"].(string)
					tamano, _ := parametros["tamano"].(float64)

					utils.InfoLog.Info(fmt.Sprintf("## (%d) - Procesando INIT_PROC", pcb.PID), "archivo", archivo, "tamaño", int(tamano))

					// Crear nuevo proceso
					nuevoPCB := NuevoPCB(-1, int(tamano))
					nuevoPCB.NombreArchivo = archivo

					utils.InfoLog.Info(fmt.Sprintf("## (%d) Se crea el proceso - Estado: NEW", nuevoPCB.PID))
					AgregarProcesoANew(nuevoPCB)
				}

				// El proceso que hizo INIT_PROC continúa en EXEC (no cambia de estado)
				// Incrementar PC ya que el proceso continúa ejecutando
				pcb.PC++
				utils.InfoLog.Info(fmt.Sprintf("## (%d) - PC incrementado después de INIT_PROC", pcb.PID), "nuevo_pc", pcb.PC)
				return true

			case "SYSCALL_IO":
				// Mover proceso a BLOCKED para operación de I/O
				utils.InfoLog.Info(fmt.Sprintf("## (%d) - Procesando IO", pcb.PID))
				pcb.CambiarEstado(EstadoBlocked)

				if parametros, ok := respuestaMap["parametros"].(map[string]interface{}); ok {
					dispositivo, _ := parametros["dispositivo"].(string)
					tiempo, _ := parametros["tiempo"].(float64)

					// Seleccionar dispositivo IO real usando balanceador de carga
					dispositivoReal := SeleccionarDispositivoIO(dispositivo, pcb.PID)

					// Mover proceso a BLOCKED con el dispositivo real seleccionado
					MoverProcesoABlocked(pcb, fmt.Sprintf("IO_%s", dispositivoReal))

					// Enviar solicitud IO al dispositivo real seleccionado
					go EnviarSolicitudIO(pcb, dispositivoReal, int(tiempo))
				}
				return true

			case "SYSCALL_DUMP_MEMORY":
				// Procesar dump de memoria
				utils.InfoLog.Info(fmt.Sprintf("## (%d) - Procesando DUMP_MEMORY", pcb.PID))
				pcb.CambiarEstado(EstadoBlocked)
				MoverProcesoABlocked(pcb, "DUMP_MEMORY")
				return true

			case "EXIT":
    			utils.InfoLog.Info(fmt.Sprintf("## (%d) - Proceso solicita EXIT", pcb.PID))
    			FinalizarProceso(pcb, "EXIT")
    			utils.InfoLog.Info(fmt.Sprintf("## (%d) - RETORNANDO DESPUÉS DE EXIT", pcb.PID)) // ← AGREGAR ESTO
    			return true

			case "ERROR":
				// Error en la ejecución
				utils.ErrorLog.Error("Error en ejecución de proceso", "pid", pcb.PID)
				FinalizarProceso(pcb, "ERROR")
				return true
			}
		}

		// Si no hay motivo específico de retorno, el proceso continúa ejecutando
		// Solo incrementar PC si no fue actualizado por CPU (ej: para GOTO)
		if !pcActualizadoPorCPU {
			pcb.PC++
		}
		utils.InfoLog.Info(fmt.Sprintf("## (%d) - Continuando ejecución", pcb.PID), "nuevo_pc", pcb.PC)

		// El proceso sigue en EXEC, enviarlo de vuelta a la CPU para la siguiente instrucción
		// Esto se maneja en la goroutine despacharYProcesarCPU que llamará nuevamente a EnviarProcesoCPU
		return true
	}

	// Si llegamos aquí, hubo un error con la respuesta
	utils.ErrorLog.Warn("Formato de respuesta inválido de CPU", "respuesta", fmt.Sprintf("%v", respuesta))
	return false
}
