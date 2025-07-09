package main

import (
	"fmt"
	"time"

	"github.com/sisoputnfrba/tp-2025-1c-LosCuervosXeneizes/utils"
)

// KernelConfig define la configuración del módulo Kernel
type KernelConfig struct {
	IPKernel               string  `json:"IP_KERNEL"`
	PortKernel             int     `json:"PUERTO_KERNEL"`
	IPMemory               string  `json:"IP_MEMORIA"`
	PortMemory             int     `json:"PUERTO_MEMORIA"`
	LogLevel               string  `json:"LOG_LEVEL"`
	SchedulerAlgorithm     string  `json:"ALGORITMO_CORTO_PLAZO"`     // Algoritmo STS (FIFO, SJF, SRT)
	ReadyIngressAlgorithm  string  `json:"ALGORITMO_INGRESO_A_READY"` // Algoritmo LTS (FIFO, PMCP)
	Alpha                  float64 `json:"ALFA"`                      // Para estimación SJF/SRT (0.0-1.0)
	InitialEstimate        int     `json:"ESTIMACION_INICIAL"`        // Estimación inicial en ms
	SuspensionTime         int     `json:"TIEMPO_SUSPENSION"`         // Tiempo para suspensión en ms
	GradoMultiprogramacion int     `json:"GRADO_MULTIPROGRAMACION"`   // Máximo procesos en memoria (opcional)
	ScriptsPath            string  `json:"SCRIPTS_PATH,omitempty"`    // Path para scripts (opcional)
}

var (
	kernelModulo  *utils.Modulo
	kernelConfig  *KernelConfig
	memoriaClient *utils.HTTPClient
	cpuClients    map[string]*utils.HTTPClient
)

// inicializarKernel optimizado
func inicializarKernel(configPath string) error {
	kernelModulo = utils.NuevoModulo("Kernel", configPath)
	kernelConfig = utils.CargarConfiguracion[KernelConfig](configPath)

	utils.InicializarLogger(kernelConfig.LogLevel, "Kernel")
	utils.InfoLog.Info("Inicializando Kernel...")

	InicializarPlanificador(kernelConfig)

	// Inicializar y conectar con Memoria
	memoriaClient = utils.NewHTTPClient(kernelConfig.IPMemory, kernelConfig.PortMemory, "Kernel->Memoria")
	if err := conectarAMemoria(10); err != nil { // 10 intentos
		utils.ErrorLog.Error("No se pudo conectar con Memoria. Abortando.", "error", err)
		return err
	}

	cpuClients = make(map[string]*utils.HTTPClient)

	registrarHandlers()
	kernelModulo.IniciarServidor(kernelConfig.IPKernel, kernelConfig.PortKernel)

	utils.InfoLog.Info("Kernel inicializado correctamente.")
	return nil
}

// registrarHandlers registra todos los manejadores HTTP
func registrarHandlers() {
	// Registrar manejadores en el módulo
	kernelModulo.RegistrarHandler(fmt.Sprintf("%d", utils.MensajeHandshake), "handshake", HandlerHandshake)
	kernelModulo.RegistrarHandler(fmt.Sprintf("%d", utils.MensajeOperacion), "default", HandlerOperacion)
}

// conectarAMemoria intenta conectar con el módulo de Memoria con reintentos
func conectarAMemoria(intentosMax int) error {
	utils.InfoLog.Info("Conectando con módulo", "destino", "Memoria", "intentos_max", intentosMax)
	for i := 0; i < intentosMax; i++ {
		err := memoriaClient.VerificarConexion()
		if err == nil {
			utils.InfoLog.Info("Conexión establecida con Memoria")
			return nil
		}
		utils.InfoLog.Warn("Fallo al conectar con Memoria, reintentando...", "intento", i+1, "error", err)
		time.Sleep(3 * time.Second) // Esperar 3 segundos antes de reintentar
	}
	return fmt.Errorf("no se pudo establecer conexión después de %d intentos", intentosMax)
}

// crearYAdmitirProcesoInicial crea el PCB inicial y lo coloca en NEW
func crearYAdmitirProcesoInicial(nombreArchivo string, tamanio int) {
	utils.InfoLog.Info("Creando proceso inicial", "archivo", nombreArchivo, "tamaño", tamanio)
	// El proceso inicial debe tener PID 0 según enunciado
	pcb := NuevoPCB(-1, tamanio) // Usar -1 para forzar que use GenerarNuevoPID() que ahora devuelve 0
	pcb.NombreArchivo = nombreArchivo

	utils.InfoLog.Info(fmt.Sprintf("## (%d) Se crea el proceso - Estado: NEW", pcb.PID))
	AgregarProcesoANew(pcb)
}

// iniciarPlanificadores se llama después de presionar Enter
func iniciarPlanificadores() {
	utils.InfoLog.Info("Iniciando planificadores...")
	go PlanificarLargoPlazo()
	go PlanificarCortoPlazo()
	utils.InfoLog.Info("Planificadores iniciados.")
}
