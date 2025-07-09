package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/sisoputnfrba/tp-2025-1c-LosCuervosXeneizes/utils"
)

func main() {
	// Inicializar loggers antes de cualquier log
	utils.InicializarLogger("INFO", "kernel")

	// Path de configuración por defecto
	rutaConfig := filepath.Join("configs", "kernel-config.json")

	// Determinar path de configuración
	configPath := rutaConfig

	utils.InfoLog.Info("Argumentos recibidos", "args", os.Args)

	if len(os.Args) > 1 && (os.Args[1] == "-c" || os.Args[1] == "--config") && len(os.Args) > 2 {
		configPath = os.Args[2]
		os.Args = append(os.Args[:1], os.Args[3:]...)
	}

	// Verificar argumentos para proceso inicial
	if len(os.Args) < 3 {
		utils.ErrorLog.Error("Uso: ./kernel [-c config_path] <archivo_pseudocódigo> <tamaño>")
		os.Exit(1)
	}

	nombreArchivoInicial := os.Args[1]
	tamanioInicial, err := strconv.Atoi(os.Args[2])
	if err != nil {
		utils.ErrorLog.Error("El tamaño del proceso inicial debe ser un número entero", "error", err)
		os.Exit(1)
	}

	// Inicializar kernel
	err = inicializarKernel(configPath)
	if err != nil {
		utils.ErrorLog.Error("Error durante la inicialización del Kernel", "error", err)
		os.Exit(1)
	}

	// Crear proceso inicial
	crearYAdmitirProcesoInicial(nombreArchivoInicial, tamanioInicial)

	utils.InfoLog.Info("Kernel listo y esperando conexiones/operaciones.")

	// REQUISITO DEL ENUNCIADO: Esperar Enter para iniciar planificadores
	fmt.Println("Presione ENTER para iniciar los planificadores...")
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	utils.InfoLog.Info("Enter presionado, iniciando planificadores de largo y corto plazo...")
	iniciarPlanificadores()

	select {}
}
