package main

import (
	"fmt"

	"github.com/sisoputnfrba/tp-2025-1c-LosCuervosXeneizes/utils"
)

type Configuracion struct {
	PageSize int
}

type TablaPagina struct {
	// Definición de la estructura de la tabla de páginas
}

// Asigna un marco libre para un proceso
func asignarMarco(pid int) (int, error) {
	// Buscar un marco libre
	for i, libre := range marcosLibres {
		if libre {
			// Marcar el marco como ocupado
			marcosLibres[i] = false

			// Registrar que este marco está asignado al proceso
			marcosAsignadosPorProceso[pid] = append(marcosAsignadosPorProceso[pid], i)

			utils.InfoLog.Info("Marco asignado", "PID", pid, "marco", i)
			return i, nil
		}
	}

	return 0, fmt.Errorf("no hay marcos libres disponibles")
}

// Cuenta el número de marcos libres disponibles
func contarMarcosLibres() int {
	count := 0
	for _, libre := range marcosLibres {
		if libre {
			count++
		}
	}
	return count
}

// liberarMemoriaProceso libera todos los marcos asignados a un proceso
func liberarMemoriaProceso(pid int) error {
	// Verificar si existe el proceso
	marcos, existe := marcosAsignadosPorProceso[pid]
	if !existe {
		return fmt.Errorf("no existe asignación de memoria para el proceso %d", pid)
	}

	// Marcar como libres todos los marcos asignados al proceso
	for _, marco := range marcos {
		marcosLibres[marco] = true

		// Opcional: limpiar la memoria (poner en ceros)
		inicio := marco * config.PageSize
		fin := inicio + config.PageSize
		for i := inicio; i < fin && i < len(memoriaPrincipal); i++ {
			memoriaPrincipal[i] = 0
		}
	}

	// Eliminar la entrada del proceso del mapa de asignaciones
	delete(marcosAsignadosPorProceso, pid)

	// Eliminar la tabla de páginas del proceso
	delete(tablasPaginas, pid)

	utils.InfoLog.Info("Memoria liberada", "PID", pid, "marcos_liberados", len(marcos))

	return nil
}
