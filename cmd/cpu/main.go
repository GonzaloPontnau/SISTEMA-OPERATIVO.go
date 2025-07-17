package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/sisoputnfrba/tp-2025-1c-LosCuervosXeneizes/utils"
)

var (
	modulo        *utils.Modulo
	identificador string // Identificador de la CPU
)

func main() {
	// Verificar argumentos
	if len(os.Args) < 2 {
		fmt.Println("Error: Uso: ./cpu [identificador] [archivo_config_opcional]")
		os.Exit(1)
	}

	identificador = os.Args[1]

	// PRIMERO inicializar módulo (que inicializa el logger)
	inicializarModulo()

	// DESPUÉS usar el logger
	utils.InfoLog.Info("CPU iniciada", "identificador", identificador)

	// Inicializar componentes de la CPU
	inicializarCPU()

	// Bloquear ejecución (mantener vivo el proceso)
	select {}
}

func inicializarModulo() {
	// Determinar archivo de configuración (SIN usar logger todavía)
	var rutaConfig string
	if len(os.Args) >= 3 {
		// Archivo de configuración especificado como parámetro
		rutaConfig = os.Args[2]
		fmt.Printf("Usando archivo de configuración específico: %s\n", rutaConfig)
	} else {
		// Archivo de configuración por defecto
		rutaConfig = filepath.Join("configs", "cpu-config.json")
		fmt.Printf("Usando archivo de configuración por defecto: %s\n", rutaConfig)
	}

	// Verificar que el archivo existe (SIN usar logger todavía)
	if _, err := os.Stat(rutaConfig); os.IsNotExist(err) {
		fmt.Printf("Error: El archivo de configuración '%s' no existe\n", rutaConfig)
		os.Exit(1)
	}

	// Crear módulo
	modulo = utils.NuevoModulo("CPU", rutaConfig)

	// PASO CRÍTICO: Inicializar logger ANTES de usarlo
	loggerName := fmt.Sprintf("CPU-%s", identificador)
	utils.InicializarLogger("INFO", loggerName)

	// AHORA SÍ puedes usar el logger
	utils.InfoLog.Info("Logger inicializado correctamente", "modulo", loggerName)
	utils.InfoLog.Info("Archivo de configuración verificado", "ruta", rutaConfig)

	// Cargar configuración
	config = utils.CargarConfiguracion[CPUConfig](rutaConfig)

	// Actualizar nivel de log con el de la configuración
	utils.InicializarLogger(config.LogLevel, loggerName)
	utils.InfoLog.Info("Nivel de log actualizado", "nivel", config.LogLevel)

	// Datos para el handshake
	datosHandshake := map[string]interface{}{
		"nombre":        "CPU",
		"tipo":          "CPU",
		"ip":            config.IPCPU,
		"puerto":        config.PortCPU,
		"identificador": identificador,
	}

	utils.InfoLog.Info("Datos de handshake preparados", "datos", fmt.Sprintf("%v", datosHandshake))

	// Registrar handlers
	RegistrarHandlers()

	// Iniciar servidor
	modulo.IniciarServidor(config.IPCPU, config.PortCPU)

	// Crear y conectar cliente a Kernel
	modulo.CrearCliente("Kernel", config.IPKernel, config.PortKernel)
	go modulo.ConectarCliente("Kernel", 2, datosHandshake)

	// Crear y conectar cliente a Memoria
	modulo.CrearCliente("Memoria", config.IPMemory, config.PortMemory)
	go modulo.ConectarCliente("Memoria", 2, datosHandshake)
}
