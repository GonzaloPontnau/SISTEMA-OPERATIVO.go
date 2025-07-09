package utils

import (
	"time"
)

// AplicarRetardo aplica un retardo simulado y lo registra
func AplicarRetardo(operacion string, duracionMs int) {
	InfoLog.Info("Aplicando retardo", "operación", operacion, "duración_ms", duracionMs)
	time.Sleep(time.Duration(duracionMs) * time.Millisecond)
	InfoLog.Info("Retardo completado", "operación", operacion)
}

// ExtraerRetardo extrae el retardo de una operación del mensaje
func ExtraerRetardo(msg *Mensaje, valorPorDefecto int) int {
	if datosMap, ok := msg.Datos.(map[string]interface{}); ok {
		if _, ok := datosMap["retardo"].(float64); ok {
			if datosMap, ok := msg.Datos.(map[string]interface{}); ok {
				if retardo, ok := datosMap["retardo"].(float64); ok {
					return int(retardo)
				}
			}
			return valorPorDefecto
		}
	}
	return valorPorDefecto
}

// ObtenerTipoOperacion obtiene el tipo de operación del mensaje
func ObtenerTipoOperacion(msg *Mensaje, valorPorDefecto string) string {
	// Intentar obtener el tipo de operación
	if datosMap, ok := msg.Datos.(map[string]interface{}); ok {
		if tipo, ok := datosMap["tipo"].(string); ok {
			return tipo
		}
	}
	return valorPorDefecto
}

// HandlerGenerico es un handler genérico para tratar operaciones con retardo
func HandlerGenerico(msg *Mensaje, retardoPorDefecto int, procesador func(msg *Mensaje) (interface{}, error)) (interface{}, error) {
	InfoLog.Info("Operación recibida", "origen", msg.Origen, "tipo", msg.Tipo)

	// Extraer retardo
	retardo := ExtraerRetardo(msg, retardoPorDefecto)

	// Aplicar retardo
	AplicarRetardo("procesamiento", retardo)

	// Procesar la operación
	return procesador(msg)
}
