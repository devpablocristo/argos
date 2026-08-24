# Argos — instrucciones del repositorio

## 1. Propósito y versión activa

Argos es un sistema agrícola de visión para datasets multiespectrales, actualmente orientado a grupos de captura DJI Mavic 3 Multispectral. El monorepo raíz es la implementación activa.

## 2. Fuentes de verdad y límites

- Leer `PROJECT_CONTEXT.md`, `README.md`, `docs/` y el código afectado según la tarea.
- Preservar reglas y decisiones útiles; mover dudas a “posiblemente obsoleto” antes de eliminarlas.
- No inventar reglas, asociaciones semánticas ni georreferenciación.

## 3. Arquitectura y estructura

- `core/`: API Go, orquestación, persistencia, migraciones e integraciones.
- `processing/`: paquete/worker Python para lectura, metadata, validación, NDVI y exports.
- `ui/`: frontend React + TypeScript.
- `docs/`: contratos, arquitectura, ADRs y roadmap.
- `sample/`: grupo de captura DJI local.

## 4. Comandos

- Runtime: `make up`, `make down`, `make run-core-dev`, `make run-ui`.
- Checks: `make core-test`, `make processing-test`, `make ui-build`, `make test`.
- Datos: `make process-sample`, `make seed-nexus-rules`, `make seed-companion-assist`.

## 5. Entornos e integraciones

- UI Docker `http://localhost:13003`; API `http://localhost:18090`; PostGIS host `15436`.
- Axis/Nexus y Companion son integraciones opcionales. Argos debe conservar el análisis NDVI y mostrar sincronización degradada si no están disponibles.
- Para desarrollo integrado, iniciar Axis antes de Argos y ejecutar los seeds documentados.

## 6. Validaciones obligatorias

Ejecutar el check más estrecho durante el cambio y `make test` al cerrar una modificación transversal. Validar migraciones y ciclo de datos cuando cambie persistencia.

## 7. Riesgos y trampas

- Dataset no equivale automáticamente a field, lot, campaign, flight o capture.
- Las clasificaciones pueden ser nulas; Argos no inventa asociaciones.
- Delete elimina filas y outputs generados por Argos, nunca imágenes fuente.
- `make down` conserva volúmenes.
- Las afirmaciones antiguas sobre Companion pueden estar desactualizadas: contrastar código y docs.

## 8. Desviaciones

Grasp está habilitado sólo como piloto consultivo de Argos. Sus scores no son gates ni sustituyen tests, Codebase Memory o revisión del código.

## 9. Herramientas y contexto bajo demanda

- Detalle preservado: `docs/agents/repository-context.md`.
- Piloto Grasp: `docs/tooling/grasp-pilot.md`.
- C4 compartido: `/home/pablocristo/.config/agent-observability/architecture/docs/c4/workspace.dsl`; cambiarlo sólo si cambian límites.
- Usar Codebase Memory para navegación, Chrome DevTools para UI real y GitButler para escrituras Git.
