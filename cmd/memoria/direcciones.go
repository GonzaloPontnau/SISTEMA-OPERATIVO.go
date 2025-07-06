package main

import (
	"fmt"
)

// Traduce una dirección lógica a una dirección física
func traducirDireccion(pid int, dirLogica int) (int, error) {
	// Obtener tabla de páginas de nivel 1 para el proceso
	tabla, existe := tablasPaginas[pid]
	if !existe {
		return 0, fmt.Errorf("no existe tabla de páginas para PID %d", pid)
	}

	// Calcular componentes de la dirección lógica
	numPagina := dirLogica / config.PageSize
	desplazamiento := dirLogica % config.PageSize

	// Obtener marco mediante función recursiva que navegue los niveles
	marco, err := obtenerMarcoDesdeTabla(pid, tabla, numPagina, 1)
	if err != nil {
		return 0, err
	}

	// Calcular dirección física
	dirFisica := marco*config.PageSize + desplazamiento

	return dirFisica, nil
}

// Calcula el número de páginas necesarias para un tamaño dado
func calcularNumeroPaginas(tamanio int) int {
	return (tamanio + config.PageSize - 1) / config.PageSize
}
