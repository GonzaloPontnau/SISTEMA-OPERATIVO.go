package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Configuración para los módulo

type IOConfig struct {
	IPKernel   string `json:"ip_kernel"`
	PortKernel int    `json:"port_kernel"`
	PortIO     int    `json:"port_io"`
	IPIO       string `json:"ip_io"`
	LogLevel   string `json:"log_level"`
}
type KernelConfig struct {
	PortCPU          int    `json:"port_cpu"`
	IPCPU            string `json:"ip_cpu"`
	IPMemory         string `json:"ip_memory"`
	PortMemory       int    `json:"port_memory"`
	IPKernel         string `json:"ip_kernel"`
	PortKernel       int    `json:"port_kernel"`
	TLBEntries       int    `json:"tlb_entries"`
	TLBReplacement   string `json:"tlb_replacement"`
	CacheEntries     int    `json:"cache_entries"`
	CacheReplacement string `json:"cache_replacement"`
	CacheDelay       int    `json:"cache_delay"`
	LogLevel         string `json:"log_level"`
	ScriptsPath      string `json:"scripts_path"`
}
type MemoryConfig struct {
	PortMemory     int    `json:"port_memory"`
	IPMemory       string `json:"ip_memory"`
	MemorySize     int    `json:"memory_size"`
	PageSize       int    `json:"page_size"`
	EntriesPerPage int    `json:"entries_per_page"`
	NumberOfLevels int    `json:"number_of_levels"`
	MemoryDelay    int    `json:"memory_delay"`
	SwapfilePath   string `json:"swapfile_path"`
	SwapDelay      int    `json:"swap_delay"`
	LogLevel       string `json:"log_level"`
	DumpPath       string `json:"dump_path"`
	ScriptsPath    string `json:"scripts_path"`
}

// LoadConfig es una función genérica para cargar cualquier tipo de configuración
// usando genéricos de Go. Se usa como: LoadConfig[CPUConfig]("ruta/al/archivo")
func LoadConfig[T any](configPath string) (*T, error) {
	var config T
	if err := loadConfig(configPath, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

// loadConfig carga un archivo de configuración JSON
func loadConfig(configPath string, config interface{}) error {
	absPath, err := filepath.Abs(configPath)
	if err != nil {
		return fmt.Errorf("error obteniendo ruta absoluta: %v", err)
	}

	file, err := os.Open(absPath)
	if err != nil {
		return fmt.Errorf("error abriendo archivo de configuración %s: %v", absPath, err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(config); err != nil {
		return fmt.Errorf("error decodificando configuración: %v", err)
	}

	return nil
}
