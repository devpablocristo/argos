# Piloto local de Grasp

Grasp 3.21.0 se usa en Argos como una segunda opinión visual y heurística. No
reemplaza a codebase-memory, al código ni a los tests, y su nota no es un gate.

## Uso

```bash
agent-observe open grasp
agent-observe grasp-refresh
agent-observe grasp-report
```

La UI está disponible sólo en `http://127.0.0.1:18081`. Codex y Claude reciben
por MCP cinco consultas de sólo lectura sobre el último informe: resumen,
hallazgos, contexto de archivo, búsqueda y procedencia. Los agentes mantienen
sus permisos normales; el alcance reducido corresponde únicamente al producto
Grasp del piloto.

Los informes quedan fuera del repositorio en
`~/.local/share/agent-observability/reports/grasp/argos`. Cada actualización
conserva una ejecución histórica y renueva la copia mostrada por la UI.

## Lectura del primer informe

El primer snapshot encontró 85 archivos, 372 funciones y 207 conexiones, con
salud heurística 65/100. También marcó nueve hallazgos. Al menos tres requieren
especial cautela antes de actuar:

- identificó valores de organización de tests como secretos;
- marcó un archivo TSX como riesgo de inyección SQL sin aportar evidencia;
- reportó ciclos Go que incluyen archivos de test y pueden ser artefactos de su
  resolución de imports.

Esto confirma que Grasp sirve para formular preguntas y recorrer el mapa, no
para decidir cambios automáticamente. Además, la UI normaliza el mismo informe
y muestra 77/100 mientras el CLI informa 65/100; las notas son orientativas y
no comparables sin conservar el mismo modo de cálculo.

## Aislamiento y procedencia

- CLI e imagen fijados en `v3.21.0` y digest
  `sha256:5f08b2b9bb5bc93665f0b78450b3c07566da0234f8160a748b52d06a0d2bab4b`.
- El análisis puntual corre sin red y monta Argos como sólo lectura.
- La UI usa dependencias web locales y una política que sólo permite conexiones
  al mismo origen.
- No se ejecuta `grasp setup`, no se instalan hooks y no se habilitan gates.
- El MCP oficial no se ejecuta: la distribución probada omite Kuzu y su árbol
  npm reportó nueve vulnerabilidades sin arreglo disponible, dos críticas. El
  puente local sólo publica el JSON ya generado.
- Grasp usa Elastic License 2.0.
