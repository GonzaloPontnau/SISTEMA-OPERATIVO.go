package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/sisoputnfrba/tp-2025-1c-LosCuervosXeneizes/utils"
)

func inicializarMemoria() {
	utils.InfoLog.Info("Inicializando memoria",
		"tamaño", config.MemorySize,
		"tamaño_página", config.PageSize,
		"niveles", config.NumberOfLevels)

	// Inicializar la memoria principal
	memoriaPrincipal = make([]byte, config.MemorySize)
	utils.InfoLog.Info("Memoria principal inicializada", "tamaño", len(memoriaPrincipal))

	// Inicializar tabla de páginas
	tablasPaginas = make(map[int]*TablaPaginas)

	// Inicializar array para rastrear marcos libres
	totalMarcos := config.MemorySize / config.PageSize
	marcosLibres = make([]bool, totalMarcos)
	for i := range marcosLibres {
		marcosLibres[i] = true // Inicialmente, todos los marcos están libres
	}

	// Inicializar el mapeo de marcos por proceso
	marcosAsignadosPorProceso = make(map[int][]int)

	// Inicializar área de swap
	utils.InfoLog.Info("Inicializando área de swap", "ruta", config.SwapfilePath)
	err := inicializarAreaSwap()
	if err != nil {
		utils.ErrorLog.Error("Error al inicializar el área de swap", "error", err)
		os.Exit(1)
	}
}

// Función para inicializar el área de swap
func inicializarAreaSwap() error {
	// Crear directorio si no existe
	dir := filepath.Dir(config.SwapfilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("error al crear directorio para swap: %v", err)
	}

	// Crear o truncar el archivo SWAP
	swapFile, err := os.Create(config.SwapfilePath)
	if err != nil {
		return fmt.Errorf("error al crear archivo SWAP: %v", err)
	}
	defer swapFile.Close()

	// Inicializar el mapa de SWAP
	mapaSwap = make(map[string]EntradaSwap)

	utils.InfoLog.Info("Archivo de SWAP inicializado", "ruta", config.SwapfilePath)
	return nil
}

// Inicializar métricas
func inicializarMetricas() {
	metricasPorProceso = make(map[int]*MetricasProceso)
}
