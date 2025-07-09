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
		fmt.Println("Error: Uso: ./cpu [identificador]")
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
	rutaConfig := filepath.Join("configs", "cpu-config.json")

	// Crear módulo
	modulo = utils.NuevoModulo("CPU", rutaConfig)

	// Inicializar logger con identificador
	loggerName := fmt.Sprintf("CPU-%s", identificador)
	utils.InicializarLogger("INFO", loggerName)

	// Cargar configuración
	config = utils.CargarConfiguracion[CPUConfig](rutaConfig)

	// Datos para el handshake
	datosHandshake := map[string]interface{}{
		"nombre":        "CPU",
		"tipo":          "CPU",
		"ip":            config.IPCPU,
		"puerto":        config.PortCPU,
		"identificador": identificador,
	}

	utils.InfoLog.Info("Datos de handshake preparados", "datos", fmt.Sprintf("%v", datosHandshake))

	// Actualizar nivel de log con el de la configuración
	utils.InicializarLogger(config.LogLevel, loggerName)

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
