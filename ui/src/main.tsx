import React, { useEffect, useMemo, useRef, useState } from 'react';
import { createRoot } from 'react-dom/client';
import './styles.css';

const apiBase = import.meta.env.VITE_ARGOS_API_URL ?? 'http://localhost:18090';
const defaultSourceURI = import.meta.env.VITE_ARGOS_DEFAULT_SOURCE_URI ?? '../sample';
const languageStorageKey = 'argos.language';

type Language = 'es' | 'en';
type ImageMode = 'rgb' | 'ndvi';

const translations = {
  es: {
    tagline: 'Pipeline de vision multiespectral con drones',
    languageLabel: 'Idioma',
    sampleDatasetName: 'Muestra DJI M3M',
    uploadedDatasetName: 'Imagenes cargadas',
    scan: 'Escanear',
    chooseFolder: 'Elegir carpeta',
    chooseImages: 'Elegir imagenes',
    processing: 'Procesando',
    fields: 'Campos',
    fieldHistory: 'Campos',
    activeFields: 'Activos',
    archivedFields: 'Archivados',
    createField: 'Crear campo',
    newFieldName: 'Nombre del campo',
    fieldNotes: 'Notas',
    noFields: 'Todavia no hay campos. Crea uno para empezar a cargar datasets.',
    noArchivedFields: 'No hay campos archivados.',
    chooseFieldHint: 'Crea o elegi un campo antes de cargar imagenes.',
    fieldDatasets: 'Datasets del campo',
    noFieldDatasets: 'Este campo todavia no tiene datasets.',
    datasetRequiresField: 'Primero crea o elegi un campo.',
    fieldArchiveConfirm: 'Archivar este campo? Sus datasets se conservan y quedan asociados.',
    fieldRestoreConfirm: 'Restaurar este campo?',
    fieldDeleteConfirm: 'Borrar fisicamente este campo? Solo se permite si esta archivado y no tiene datasets.',
    dataset: 'Dataset',
    name: 'Nombre',
    status: 'Estado',
    source: 'Origen',
    createDatasetHint: 'Crea un dataset desde la ruta local de muestra.',
    history: 'Historial global de datasets',
    datasetHistory: 'Historial del dataset',
    semanticClassification: 'Clasificacion semantica',
    datasetType: 'Tipo',
    scope: 'Alcance',
    confidence: 'Confianza',
    missingMetadata: 'Metadata faltante',
    noMissingMetadata: 'Completa',
    field: 'Campo',
    noFieldAssociated: 'Sin campo asociado',
    noDatasetEvents: 'Sin eventos para este dataset.',
    activeDatasets: 'Activos',
    archivedDatasets: 'Archivados',
    noDatasets: 'Todavia no hay datasets guardados.',
    noArchivedDatasets: 'No hay datasets archivados.',
    createdAt: 'Creado',
    archivedAt: 'Archivado',
    update: 'Actualizar',
    archive: 'Archivar',
    restore: 'Restaurar',
    hardDelete: 'Borrar fisico',
    archiveConfirm: 'Archivar este dataset? Lo oculta del historial activo, pero conserva capturas y analisis.',
    restoreConfirm: 'Restaurar este dataset al historial activo?',
    deleteConfirm: 'Borrar fisicamente este dataset de Argos? Se eliminan sus capturas, analisis y outputs generados. Las imagenes fuente no se borran.',
    captures: 'Capturas',
    noCaptures: 'Todavia no hay capturas escaneadas.',
    analysis: 'Analisis',
    rgb: 'RGB',
    ndvi: 'NDVI',
    zoomLabel: 'Zoom de imagen',
    fit: 'Ajustar',
    findings: 'Hallazgos',
    noFindings: 'Sin hallazgos para este analisis.',
    assistedInterpretation: 'Interpretacion asistida',
    noInterpretation: 'Sin interpretacion asistida todavia.',
    unavailable: 'No disponible',
    syncStatuses: {
      pending: 'pendiente',
      synced: 'sincronizado',
      failed: 'fallo',
      not_configured: 'no configurado',
    },
    assist: {
      summary: 'Resumen',
      simple_explanation: 'Explicacion simple',
      limitations: 'Limitaciones',
      suggested_questions: 'Preguntas sugeridas',
      next_steps: 'Proximos pasos',
      raw_text: 'Texto',
    },
    unexpectedError: 'Error inesperado',
    imageAlt: 'Preview NDVI',
    rgbAlt: 'Imagen RGB',
    metrics: {
      mean: 'Promedio',
      min: 'Min',
      max: 'Max',
      std: 'Desvio',
      pixels: 'Pixeles',
    },
    vigor: {
      nonVegetation: 'Sin vegetacion',
      low: 'Bajo vigor',
      medium: 'Vigor medio',
      high: 'Alto vigor',
    },
    statuses: {
      registered: 'registrado',
      processing: 'procesando',
      processed: 'procesado',
      failed: 'fallido',
      valid: 'valida',
      invalid: 'invalida',
      completed: 'completado',
    },
    outputs: {
      preview_png: 'Preview PNG',
      analysis_tiff: 'Analisis TIFF',
      metadata_json: 'Metadata JSON',
    },
    datasetTypes: {
      sample: 'Sample',
      uploaded_folder: 'Carpeta subida',
      flight_dataset: 'Dataset de vuelo',
      single_capture: 'Captura unica',
      multi_capture_dataset: 'Multiples capturas',
      sector_capture: 'Captura de sector',
      unknown: 'Desconocido',
    },
    scopes: {
      global: 'Global',
      field: 'Campo',
      lot: 'Lote',
      campaign: 'Campaña',
      flight: 'Vuelo',
      dataset: 'Dataset',
    },
    eventStatuses: {
      pending: 'pendiente',
      running: 'en curso',
      completed: 'completado',
      failed: 'fallido',
    },
  },
  en: {
    tagline: 'Multispectral drone vision pipeline',
    languageLabel: 'Language',
    sampleDatasetName: 'Sample DJI M3M',
    uploadedDatasetName: 'Uploaded images',
    scan: 'Scan',
    chooseFolder: 'Choose folder',
    chooseImages: 'Choose images',
    processing: 'Processing',
    fields: 'Fields',
    fieldHistory: 'Fields',
    activeFields: 'Active',
    archivedFields: 'Archived',
    createField: 'Create field',
    newFieldName: 'Field name',
    fieldNotes: 'Notes',
    noFields: 'No fields yet. Create one before loading datasets.',
    noArchivedFields: 'No archived fields yet.',
    chooseFieldHint: 'Create or choose a field before loading images.',
    fieldDatasets: 'Field datasets',
    noFieldDatasets: 'This field does not have datasets yet.',
    datasetRequiresField: 'Create or choose a field first.',
    fieldArchiveConfirm: 'Archive this field? Its datasets are preserved and remain associated.',
    fieldRestoreConfirm: 'Restore this field?',
    fieldDeleteConfirm: 'Hard delete this field? Only allowed when archived and it has no datasets.',
    dataset: 'Dataset',
    name: 'Name',
    status: 'Status',
    source: 'Source',
    createDatasetHint: 'Create a dataset from the local sample path.',
    history: 'Global dataset history',
    datasetHistory: 'Dataset history',
    semanticClassification: 'Semantic classification',
    datasetType: 'Type',
    scope: 'Scope',
    confidence: 'Confidence',
    missingMetadata: 'Missing metadata',
    noMissingMetadata: 'Complete',
    field: 'Field',
    noFieldAssociated: 'No field associated',
    noDatasetEvents: 'No events for this dataset.',
    activeDatasets: 'Active',
    archivedDatasets: 'Archived',
    noDatasets: 'No saved datasets yet.',
    noArchivedDatasets: 'No archived datasets yet.',
    createdAt: 'Created',
    archivedAt: 'Archived',
    update: 'Update',
    archive: 'Archive',
    restore: 'Restore',
    hardDelete: 'Hard delete',
    archiveConfirm: 'Archive this dataset? It leaves captures and analyses intact, but hides it from active history.',
    restoreConfirm: 'Restore this dataset to active history?',
    deleteConfirm: 'Hard delete this dataset from Argos? Its captures, analyses and generated outputs are removed. Source images are not deleted.',
    captures: 'Captures',
    noCaptures: 'No captures scanned yet.',
    analysis: 'Analysis',
    rgb: 'RGB',
    ndvi: 'NDVI',
    zoomLabel: 'Image zoom',
    fit: 'Fit',
    findings: 'Findings',
    noFindings: 'No findings for this analysis.',
    assistedInterpretation: 'Assisted interpretation',
    noInterpretation: 'No assisted interpretation yet.',
    unavailable: 'Unavailable',
    syncStatuses: {
      pending: 'pending',
      synced: 'synced',
      failed: 'failed',
      not_configured: 'not configured',
    },
    assist: {
      summary: 'Summary',
      simple_explanation: 'Simple explanation',
      limitations: 'Limitations',
      suggested_questions: 'Suggested questions',
      next_steps: 'Next steps',
      raw_text: 'Text',
    },
    unexpectedError: 'Unexpected error',
    imageAlt: 'NDVI preview',
    rgbAlt: 'RGB image',
    metrics: {
      mean: 'Mean',
      min: 'Min',
      max: 'Max',
      std: 'Std',
      pixels: 'Pixels',
    },
    vigor: {
      nonVegetation: 'No vegetation',
      low: 'Low vigor',
      medium: 'Medium vigor',
      high: 'High vigor',
    },
    statuses: {
      registered: 'registered',
      processing: 'processing',
      processed: 'processed',
      failed: 'failed',
      valid: 'valid',
      invalid: 'invalid',
      completed: 'completed',
    },
    outputs: {
      preview_png: 'Preview PNG',
      analysis_tiff: 'Analysis TIFF',
      metadata_json: 'Metadata JSON',
    },
    datasetTypes: {
      sample: 'Sample',
      uploaded_folder: 'Uploaded folder',
      flight_dataset: 'Flight dataset',
      single_capture: 'Single capture',
      multi_capture_dataset: 'Multiple captures',
      sector_capture: 'Sector capture',
      unknown: 'Unknown',
    },
    scopes: {
      global: 'Global',
      field: 'Field',
      lot: 'Lot',
      campaign: 'Campaign',
      flight: 'Flight',
      dataset: 'Dataset',
    },
    eventStatuses: {
      pending: 'pending',
      running: 'running',
      completed: 'completed',
      failed: 'failed',
    },
  },
} as const;

type Translation = (typeof translations)[Language];

type Field = {
  id: string;
  org_id?: string;
  name: string;
  notes: string;
  created_at: string;
  updated_at: string;
  archived_at?: string | null;
};

type Dataset = {
  id: string;
  name: string;
  source_uri: string;
  status: string;
  field_id?: string | null;
  classification?: DatasetClassification;
  created_at: string;
  updated_at: string;
  archived_at?: string | null;
};

type DatasetClassification = {
  dataset_id: string;
  dataset_name: string;
  dataset_type: string;
  scope: string;
  field_id?: string | null;
  lot_id?: string | null;
  campaign_id?: string | null;
  flight_id?: string | null;
  capture_ids: string[];
  analysis_ids: string[];
  confidence: number;
  missing_metadata: string[];
  reason: string;
  classified_at: string;
};

type DatasetEvent = {
  event_id: string;
  dataset_id: string;
  event_type: string;
  timestamp: string;
  status: string;
  message: string;
  details: Record<string, unknown>;
};

type Capture = {
  id: string;
  capture_key: string;
  captured_at?: string;
  validation_status: string;
  warnings?: string[];
  errors?: string[];
  analysis?: {
    kind: string;
    status: string;
    metrics: Record<string, number>;
  };
};

type Analysis = {
  id: string;
  dataset_id?: string;
  capture_id?: string;
  kind: string;
  status: string;
  metrics: Record<string, number>;
  outputs: OutputAsset[];
  nexus_sync_status?: string;
  nexus_sync_error?: string;
  nexus_correlation_id?: string;
  nexus_findings?: FindingSnapshot[];
  companion_sync_status?: string;
  companion_sync_error?: string;
  companion_correlation_id?: string;
  companion_output?: Record<string, unknown>;
};

type FindingSnapshot = {
  id?: string;
  code: string;
  severity: string;
  title: string;
  message: string;
  recommendation?: string;
  status: string;
};

type OutputAsset = {
  id: string;
  kind: string;
  content_type: string;
  byte_size: number;
};

type UploadScanResponse = {
  dataset: Dataset;
  status: string;
  captures: Capture[];
  warnings?: string[];
};

const directoryInputAttributes = {
  webkitdirectory: '',
  directory: '',
};

async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
  const response = await fetch(`${apiBase}${path}`, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...(options.headers ?? {}),
    },
  });
  if (!response.ok) {
    const body = await response.text();
    throw new Error(body || response.statusText);
  }
  return (await response.json()) as T;
}

async function requestForm<T>(path: string, body: FormData): Promise<T> {
  const response = await fetch(`${apiBase}${path}`, {
    method: 'POST',
    body,
  });
  if (!response.ok) {
    const text = await response.text();
    throw new Error(text || response.statusText);
  }
  return (await response.json()) as T;
}

function initialLanguage(): Language {
  try {
    const stored = localStorage.getItem(languageStorageKey);
    if (stored === 'es' || stored === 'en') {
      return stored;
    }
  } catch {
    return 'es';
  }
  return navigator.language.toLowerCase().startsWith('es') ? 'es' : 'en';
}

function App() {
  const uploadInputRef = useRef<HTMLInputElement | null>(null);
  const folderInputRef = useRef<HTMLInputElement | null>(null);
  const [language, setLanguage] = useState<Language>(initialLanguage);
  const [sourceURI, setSourceURI] = useState(defaultSourceURI);
  const [fields, setFields] = useState<Field[]>([]);
  const [field, setField] = useState<Field | null>(null);
  const [showArchivedFields, setShowArchivedFields] = useState(false);
  const [newFieldName, setNewFieldName] = useState('');
  const [newFieldNotes, setNewFieldNotes] = useState('');
  const [fieldEditName, setFieldEditName] = useState('');
  const [fieldEditNotes, setFieldEditNotes] = useState('');
  const [datasets, setDatasets] = useState<Dataset[]>([]);
  const [dataset, setDataset] = useState<Dataset | null>(null);
  const [showArchived, setShowArchived] = useState(false);
  const [datasetEditName, setDatasetEditName] = useState('');
  const [datasetEditSourceURI, setDatasetEditSourceURI] = useState('');
  const [datasetClassification, setDatasetClassification] = useState<DatasetClassification | null>(null);
  const [datasetEvents, setDatasetEvents] = useState<DatasetEvent[]>([]);
  const [captures, setCaptures] = useState<Capture[]>([]);
  const [analysis, setAnalysis] = useState<Analysis | null>(null);
  const [selectedCaptureID, setSelectedCaptureID] = useState<string | null>(null);
  const [imageMode, setImageMode] = useState<ImageMode>('rgb');
  const [imageZoom, setImageZoom] = useState(100);
  const [fitToScreen, setFitToScreen] = useState(true);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const t = translations[language];

  useEffect(() => {
    document.documentElement.lang = language;
    localStorage.setItem(languageStorageKey, language);
  }, [language]);

  useEffect(() => {
    void loadInitialData();
  }, []);

  const ndviPreview = useMemo(() => {
    const output = analysis?.outputs.find((item) => item.kind === 'preview_png');
    return output ? `${apiBase}/v1/assets/${output.id}` : '';
  }, [analysis]);

  const rgbPreview = selectedCaptureID ? `${apiBase}/v1/captures/${selectedCaptureID}/assets/rgb` : '';
  const activePreview = imageMode === 'rgb' ? rgbPreview : ndviPreview;
  const outputs = useMemo(() => analysis?.outputs ?? [], [analysis]);
  const visibleDatasets = useMemo(
    () => datasets.filter((item) => (showArchived ? Boolean(item.archived_at) : !item.archived_at)),
    [datasets, showArchived],
  );
  const visibleFields = useMemo(
    () => fields.filter((item) => (showArchivedFields ? Boolean(item.archived_at) : !item.archived_at)),
    [fields, showArchivedFields],
  );

  async function loadInitialData() {
    setBusy(true);
    setError(null);
    try {
      const response = await request<{ fields: Field[] }>('/v1/fields?include_archived=true');
      const loadedFields = response.fields ?? [];
      setFields(loadedFields);
      const latest = loadedFields.find((item) => !item.archived_at) ?? loadedFields[0];
      if (!latest) {
        setDatasets([]);
        clearDatasetSelection();
        return;
      }
      setShowArchivedFields(Boolean(latest.archived_at));
      await openField(latest);
    } catch (err) {
      setError(err instanceof Error ? err.message : translations[initialLanguage()].unexpectedError);
    } finally {
      setBusy(false);
    }
  }

  async function openField(nextField: Field) {
    setField(nextField);
    setFieldEditName(nextField.name);
    setFieldEditNotes(nextField.notes ?? '');
    setShowArchivedFields(Boolean(nextField.archived_at));
    clearDatasetSelection();

    const response = await request<{ datasets: Dataset[] }>(`/v1/fields/${nextField.id}/datasets?include_archived=true`);
    const loadedDatasets = response.datasets ?? [];
    setDatasets(loadedDatasets);
    const latest = loadedDatasets.find((item) => !item.archived_at) ?? loadedDatasets[0];
    if (!latest) {
      return;
    }
    setShowArchived(Boolean(latest.archived_at));
    await openDataset(latest);
  }

  async function selectField(nextField: Field) {
    setBusy(true);
    setError(null);
    try {
      await openField(nextField);
    } catch (err) {
      setError(err instanceof Error ? err.message : t.unexpectedError);
    } finally {
      setBusy(false);
    }
  }

  function replaceField(updated: Field) {
    setFields((items) => {
      const next = items.map((item) => (item.id === updated.id ? updated : item));
      return next.some((item) => item.id === updated.id) ? next : [updated, ...next];
    });
  }

  async function createField(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setBusy(true);
    setError(null);
    try {
      const created = await request<Field>('/v1/fields', {
        method: 'POST',
        body: JSON.stringify({ name: newFieldName, notes: newFieldNotes }),
      });
      setNewFieldName('');
      setNewFieldNotes('');
      replaceField(created);
      await openField(created);
    } catch (err) {
      setError(err instanceof Error ? err.message : t.unexpectedError);
    } finally {
      setBusy(false);
    }
  }

  async function updateSelectedField(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!field) {
      return;
    }
    setBusy(true);
    setError(null);
    try {
      const updated = await request<Field>(`/v1/fields/${field.id}`, {
        method: 'PATCH',
        body: JSON.stringify({ name: fieldEditName, notes: fieldEditNotes }),
      });
      setField(updated);
      setFieldEditName(updated.name);
      setFieldEditNotes(updated.notes ?? '');
      replaceField(updated);
    } catch (err) {
      setError(err instanceof Error ? err.message : t.unexpectedError);
    } finally {
      setBusy(false);
    }
  }

  async function archiveSelectedField() {
    if (!field || !window.confirm(t.fieldArchiveConfirm)) {
      return;
    }
    setBusy(true);
    setError(null);
    try {
      const archived = await request<Field>(`/v1/fields/${field.id}/archive`, { method: 'POST' });
      setField(archived);
      setFieldEditName(archived.name);
      setFieldEditNotes(archived.notes ?? '');
      setShowArchivedFields(true);
      replaceField(archived);
    } catch (err) {
      setError(err instanceof Error ? err.message : t.unexpectedError);
    } finally {
      setBusy(false);
    }
  }

  async function restoreSelectedField() {
    if (!field || !window.confirm(t.fieldRestoreConfirm)) {
      return;
    }
    setBusy(true);
    setError(null);
    try {
      const restored = await request<Field>(`/v1/fields/${field.id}/restore`, { method: 'POST' });
      setField(restored);
      setFieldEditName(restored.name);
      setFieldEditNotes(restored.notes ?? '');
      setShowArchivedFields(false);
      replaceField(restored);
    } catch (err) {
      setError(err instanceof Error ? err.message : t.unexpectedError);
    } finally {
      setBusy(false);
    }
  }

  async function deleteSelectedField() {
    if (!field || !window.confirm(t.fieldDeleteConfirm)) {
      return;
    }
    setBusy(true);
    setError(null);
    try {
      await request<Field>(`/v1/fields/${field.id}`, { method: 'DELETE' });
      const remaining = fields.filter((item) => item.id !== field.id);
      setFields(remaining);
      const next =
        remaining.find((item) => (showArchivedFields ? Boolean(item.archived_at) : !item.archived_at)) ??
        remaining.find((item) => !item.archived_at) ??
        remaining[0];
      if (next) {
        await openField(next);
      } else {
        setField(null);
        setFieldEditName('');
        setFieldEditNotes('');
        setDatasets([]);
        clearDatasetSelection();
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : t.unexpectedError);
    } finally {
      setBusy(false);
    }
  }

  async function openDataset(nextDataset: Dataset) {
    setDataset(nextDataset);
    setSourceURI(nextDataset.source_uri);
    setDatasetEditName(nextDataset.name);
    setDatasetEditSourceURI(nextDataset.source_uri);
    setDatasetClassification(nextDataset.classification ?? null);
    setDatasetEvents([]);
    setAnalysis(null);
    setSelectedCaptureID(null);
    setImageMode('rgb');
    setImageZoom(100);
    setFitToScreen(true);

    const [captureResponse, classification, eventResponse] = await Promise.all([
      request<{ captures: Capture[] }>(`/v1/datasets/${nextDataset.id}/captures`),
      request<DatasetClassification>(`/v1/datasets/${nextDataset.id}/classification`),
      request<{ events: DatasetEvent[] }>(`/v1/datasets/${nextDataset.id}/events`),
    ]);
    setDatasetClassification(classification);
    setDatasetEvents(eventResponse.events);
    setDataset({ ...nextDataset, classification });
    replaceDataset({ ...nextDataset, classification });
    setCaptures(captureResponse.captures);
    const firstCapture = captureResponse.captures.find((capture) => capture.validation_status === 'valid') ?? captureResponse.captures[0];
    setSelectedCaptureID(firstCapture?.id ?? null);
    if (firstCapture?.validation_status === 'valid') {
      await createAnalysisForCapture(firstCapture.id);
    }
  }

  async function selectDataset(nextDataset: Dataset) {
    setBusy(true);
    setError(null);
    try {
      await openDataset(nextDataset);
    } catch (err) {
      setError(err instanceof Error ? err.message : t.unexpectedError);
    } finally {
      setBusy(false);
    }
  }

  function replaceDataset(updated: Dataset) {
    setDatasets((items) => {
      const next = items.map((item) => (item.id === updated.id ? updated : item));
      return next.some((item) => item.id === updated.id) ? next : [updated, ...next];
    });
  }

  function clearDatasetSelection() {
    setDataset(null);
    setDatasetEditName('');
    setDatasetEditSourceURI('');
    setDatasetClassification(null);
    setDatasetEvents([]);
    setCaptures([]);
    setAnalysis(null);
    setSelectedCaptureID(null);
    setImageMode('rgb');
    setImageZoom(100);
    setFitToScreen(true);
  }

  async function updateSelectedDataset(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!dataset) {
      return;
    }
    setBusy(true);
    setError(null);
    try {
      const updated = await request<Dataset>(`/v1/datasets/${dataset.id}`, {
        method: 'PATCH',
        body: JSON.stringify({ name: datasetEditName, source_uri: datasetEditSourceURI }),
      });
      setDataset(updated);
      setDatasetClassification(updated.classification ?? null);
      setSourceURI(updated.source_uri);
      setDatasetEditName(updated.name);
      setDatasetEditSourceURI(updated.source_uri);
      replaceDataset(updated);
    } catch (err) {
      setError(err instanceof Error ? err.message : t.unexpectedError);
    } finally {
      setBusy(false);
    }
  }

  async function archiveSelectedDataset() {
    if (!dataset || !window.confirm(t.archiveConfirm)) {
      return;
    }
    setBusy(true);
    setError(null);
    try {
      const archived = await request<Dataset>(`/v1/datasets/${dataset.id}/archive`, { method: 'POST' });
      setDataset(archived);
      setDatasetClassification(archived.classification ?? datasetClassification);
      setDatasetEditName(archived.name);
      setDatasetEditSourceURI(archived.source_uri);
      setShowArchived(true);
      replaceDataset(archived);
    } catch (err) {
      setError(err instanceof Error ? err.message : t.unexpectedError);
    } finally {
      setBusy(false);
    }
  }

  async function restoreSelectedDataset() {
    if (!dataset || !window.confirm(t.restoreConfirm)) {
      return;
    }
    setBusy(true);
    setError(null);
    try {
      const restored = await request<Dataset>(`/v1/datasets/${dataset.id}/restore`, { method: 'POST' });
      setDataset(restored);
      setDatasetClassification(restored.classification ?? datasetClassification);
      setDatasetEditName(restored.name);
      setDatasetEditSourceURI(restored.source_uri);
      setShowArchived(false);
      replaceDataset(restored);
    } catch (err) {
      setError(err instanceof Error ? err.message : t.unexpectedError);
    } finally {
      setBusy(false);
    }
  }

  async function deleteSelectedDataset() {
    if (!dataset || !window.confirm(t.deleteConfirm)) {
      return;
    }
    setBusy(true);
    setError(null);
    try {
      await request<Dataset>(`/v1/datasets/${dataset.id}`, { method: 'DELETE' });
      const remaining = datasets.filter((item) => item.id !== dataset.id);
      setDatasets(remaining);
      const next =
        remaining.find((item) => (showArchived ? Boolean(item.archived_at) : !item.archived_at)) ??
        remaining.find((item) => !item.archived_at) ??
        remaining[0];
      if (next) {
        setShowArchived(Boolean(next.archived_at));
        await openDataset(next);
      } else {
        clearDatasetSelection();
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : t.unexpectedError);
    } finally {
      setBusy(false);
    }
  }

  async function createAndScan() {
    if (!field) {
      setError(t.datasetRequiresField);
      return;
    }
    setBusy(true);
    setError(null);
    setAnalysis(null);
    setSelectedCaptureID(null);
    setImageMode('rgb');
    setImageZoom(100);
    setFitToScreen(true);
    try {
      const created = await request<Dataset>(`/v1/fields/${field.id}/datasets`, {
        method: 'POST',
        body: JSON.stringify({ name: t.sampleDatasetName, source_uri: sourceURI }),
      });
      const scan = await request<{ captures: Capture[] }>(`/v1/datasets/${created.id}/scan`, {
        method: 'POST',
      });
      const processedDataset = {
        ...created,
        status: scan.captures.some((capture) => capture.validation_status === 'valid') ? 'processed' : 'failed',
      };
      setDataset(processedDataset);
      setDatasetEditName(processedDataset.name);
      setDatasetEditSourceURI(processedDataset.source_uri);
      setShowArchived(false);
      setDatasets((items) => [processedDataset, ...items.filter((item) => item.id !== processedDataset.id)]);
      setCaptures(scan.captures);
      await refreshDatasetSemantics(processedDataset.id);
      const firstCapture = scan.captures.find((capture) => capture.validation_status === 'valid') ?? scan.captures[0];
      setSelectedCaptureID(firstCapture?.id ?? null);
      if (firstCapture?.validation_status === 'valid') {
        await createAnalysisForCapture(firstCapture.id);
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : t.unexpectedError);
    } finally {
      setBusy(false);
    }
  }

  async function uploadAndScan(event: React.ChangeEvent<HTMLInputElement>) {
    const files = Array.from(event.target.files ?? []);
    event.target.value = '';
    if (files.length === 0) {
      return;
    }
    if (!field) {
      setError(t.datasetRequiresField);
      return;
    }
    setBusy(true);
    setError(null);
    setAnalysis(null);
    setSelectedCaptureID(null);
    setImageMode('rgb');
    setImageZoom(100);
    setFitToScreen(true);
    try {
      const form = new FormData();
      form.append('name', inferUploadDatasetName(files, t.uploadedDatasetName));
      for (const file of files) {
        form.append('files', file, uploadFileName(file));
      }
      const scan = await requestForm<UploadScanResponse>(`/v1/fields/${field.id}/datasets/upload-scan`, form);
      const processedDataset = {
        ...scan.dataset,
        status: scan.captures.some((capture) => capture.validation_status === 'valid') ? 'processed' : 'failed',
      };
      setDataset(processedDataset);
      setSourceURI(processedDataset.source_uri);
      setDatasetEditName(processedDataset.name);
      setDatasetEditSourceURI(processedDataset.source_uri);
      setShowArchived(false);
      setDatasets((items) => [processedDataset, ...items.filter((item) => item.id !== processedDataset.id)]);
      setCaptures(scan.captures);
      await refreshDatasetSemantics(processedDataset.id);
      const firstCapture = scan.captures.find((capture) => capture.validation_status === 'valid') ?? scan.captures[0];
      setSelectedCaptureID(firstCapture?.id ?? null);
      if (firstCapture?.validation_status === 'valid') {
        await createAnalysisForCapture(firstCapture.id);
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : t.unexpectedError);
    } finally {
      setBusy(false);
    }
  }

  async function createAnalysisForCapture(captureID: string) {
    setSelectedCaptureID(captureID);
    const created = await request<Analysis>('/v1/analyses', {
      method: 'POST',
      body: JSON.stringify({ target_type: 'capture', target_id: captureID, kind: 'ndvi' }),
    });
    setAnalysis(created);
    setImageMode('ndvi');
    setImageZoom(100);
    setFitToScreen(true);
    if (created.dataset_id) {
      await refreshDatasetSemantics(created.dataset_id);
    }
  }

  async function refreshDatasetSemantics(datasetID: string) {
    const [classification, eventResponse] = await Promise.all([
      request<DatasetClassification>(`/v1/datasets/${datasetID}/classification`),
      request<{ events: DatasetEvent[] }>(`/v1/datasets/${datasetID}/events`),
    ]);
    setDatasetClassification(classification);
    setDatasetEvents(eventResponse.events);
    setDataset((current) => (current?.id === datasetID ? { ...current, classification } : current));
    setDatasets((items) => items.map((item) => (item.id === datasetID ? { ...item, classification } : item)));
  }

  async function refreshDatasetEvents(datasetID: string) {
    const response = await request<{ events: DatasetEvent[] }>(`/v1/datasets/${datasetID}/events`);
    setDatasetEvents(response.events);
  }

  async function selectCapture(capture: Capture) {
    setBusy(true);
    setError(null);
    setAnalysis(null);
    setSelectedCaptureID(capture.id);
    setImageMode('rgb');
    setImageZoom(100);
    setFitToScreen(true);
    try {
      if (capture.validation_status === 'valid') {
        await createAnalysisForCapture(capture.id);
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : t.unexpectedError);
    } finally {
      setBusy(false);
    }
  }

  function adjustZoom(delta: number) {
    setFitToScreen(false);
    setImageZoom((value) => Math.min(300, Math.max(100, value + delta)));
  }

  function resetZoom() {
    setFitToScreen(false);
    setImageZoom(100);
  }

  function selectImageMode(mode: ImageMode) {
    setImageMode(mode);
    setImageZoom(100);
    setFitToScreen(true);
  }

  return (
    <main className="app">
      <section className="toolbar">
        <div>
          <h1>Argos</h1>
          <p>{t.tagline}</p>
        </div>
        <div className="toolbar-actions">
          <div className="language-switch" aria-label={t.languageLabel}>
            <button className={language === 'es' ? 'active' : ''} aria-pressed={language === 'es'} onClick={() => setLanguage('es')}>
              ES
            </button>
            <button className={language === 'en' ? 'active' : ''} aria-pressed={language === 'en'} onClick={() => setLanguage('en')}>
              EN
            </button>
          </div>
          <div className="dataset-form">
            <input value={sourceURI} onChange={(event) => setSourceURI(event.target.value)} />
            <input ref={uploadInputRef} className="file-input" type="file" accept=".jpg,.jpeg,.tif,.tiff,.png,image/*" multiple onChange={uploadAndScan} />
            <input
              ref={folderInputRef}
              className="file-input"
              type="file"
              multiple
              {...directoryInputAttributes}
              onChange={uploadAndScan}
            />
            <button type="button" disabled={busy || !field} onClick={() => folderInputRef.current?.click()}>
              {t.chooseFolder}
            </button>
            <button type="button" className="secondary-button" disabled={busy || !field} onClick={() => uploadInputRef.current?.click()}>
              {t.chooseImages}
            </button>
            <button disabled={busy || !field} onClick={createAndScan}>
              {busy ? t.processing : t.scan}
            </button>
          </div>
        </div>
      </section>

      {error ? <pre className="error">{error}</pre> : null}

      <section className="summary-layout">
        <aside className="panel history">
          <h2>{t.fields}</h2>
          <form className="dataset-editor field-create-form" onSubmit={createField}>
            <label>
              <span>{t.newFieldName}</span>
              <input value={newFieldName} disabled={busy} onChange={(event) => setNewFieldName(event.target.value)} />
            </label>
            <label>
              <span>{t.fieldNotes}</span>
              <input value={newFieldNotes} disabled={busy} onChange={(event) => setNewFieldNotes(event.target.value)} />
            </label>
            <button type="submit" disabled={busy || !newFieldName.trim()}>
              {t.createField}
            </button>
          </form>
          <div className="history-tabs">
            <button className={!showArchivedFields ? 'active' : ''} aria-pressed={!showArchivedFields} onClick={() => setShowArchivedFields(false)}>
              {t.activeFields}
            </button>
            <button className={showArchivedFields ? 'active' : ''} aria-pressed={showArchivedFields} onClick={() => setShowArchivedFields(true)}>
              {t.archivedFields}
            </button>
          </div>
          {visibleFields.map((item) => (
            <button
              key={item.id}
              className={field?.id === item.id ? 'dataset-row selected' : 'dataset-row'}
              disabled={busy}
              onClick={() => selectField(item)}
            >
              <strong>{item.name}</strong>
              <span>{item.notes || '-'}</span>
              <small>
                {item.archived_at ? `${t.archivedAt} ${formatDate(item.archived_at, language)}` : `${t.createdAt} ${formatDate(item.created_at, language)}`}
              </small>
            </button>
          ))}
          {visibleFields.length === 0 ? <p className="muted">{showArchivedFields ? t.noArchivedFields : t.noFields}</p> : null}
        </aside>

        <aside className="panel">
          <h2>{t.field}</h2>
          {field ? (
            <>
              <dl>
                <dt>{t.name}</dt>
                <dd>{field.name}</dd>
                <dt>{t.fieldNotes}</dt>
                <dd>{field.notes || '-'}</dd>
                {field.archived_at ? (
                  <>
                    <dt>{t.archivedAt}</dt>
                    <dd>{formatDate(field.archived_at, language)}</dd>
                  </>
                ) : null}
              </dl>
              <form className="dataset-editor" onSubmit={updateSelectedField}>
                <label>
                  <span>{t.name}</span>
                  <input value={fieldEditName} disabled={busy} onChange={(event) => setFieldEditName(event.target.value)} />
                </label>
                <label>
                  <span>{t.fieldNotes}</span>
                  <input value={fieldEditNotes} disabled={busy} onChange={(event) => setFieldEditNotes(event.target.value)} />
                </label>
                <div className="dataset-actions">
                  <button type="submit" disabled={busy || !fieldEditName.trim()}>
                    {t.update}
                  </button>
                  {field.archived_at ? (
                    <button type="button" className="secondary-button" disabled={busy} onClick={restoreSelectedField}>
                      {t.restore}
                    </button>
                  ) : (
                    <button type="button" className="secondary-button" disabled={busy} onClick={archiveSelectedField}>
                      {t.archive}
                    </button>
                  )}
                  {field.archived_at ? (
                    <button type="button" className="danger-button" disabled={busy} onClick={deleteSelectedField}>
                      {t.hardDelete}
                    </button>
                  ) : null}
                </div>
              </form>
            </>
          ) : (
            <p className="muted">{t.chooseFieldHint}</p>
          )}
        </aside>

        <aside className="panel">
          <h2>{t.fieldDatasets}</h2>
          {field ? (
            <>
              <div className="history-tabs">
                <button className={!showArchived ? 'active' : ''} aria-pressed={!showArchived} onClick={() => setShowArchived(false)}>
                  {t.activeDatasets}
                </button>
                <button className={showArchived ? 'active' : ''} aria-pressed={showArchived} onClick={() => setShowArchived(true)}>
                  {t.archivedDatasets}
                </button>
              </div>
              <div className="dataset-list">
                {visibleDatasets.map((item) => (
                  <button
                    key={item.id}
                    className={dataset?.id === item.id ? 'dataset-row selected' : 'dataset-row'}
                    disabled={busy}
                    onClick={() => selectDataset(item)}
                  >
                    <strong>{item.name}</strong>
                    <span>{formatStatus(item.status, t)}</span>
                    <small>
                      {item.archived_at ? `${t.archivedAt} ${formatDate(item.archived_at, language)}` : `${t.createdAt} ${formatDate(item.created_at, language)}`}
                    </small>
                  </button>
                ))}
                {visibleDatasets.length === 0 ? <p className="muted">{showArchived ? t.noArchivedDatasets : t.noFieldDatasets}</p> : null}
              </div>
            </>
          ) : null}
          {dataset ? (
            <>
              <dl>
                <dt>{t.name}</dt>
                <dd>{dataset.name}</dd>
                <dt>{t.status}</dt>
                <dd>{formatStatus(dataset.status, t)}</dd>
                <dt>{t.source}</dt>
                <dd>{dataset.source_uri}</dd>
                {dataset.archived_at ? (
                  <>
                    <dt>{t.archivedAt}</dt>
                    <dd>{formatDate(dataset.archived_at, language)}</dd>
                  </>
                ) : null}
              </dl>
              {datasetClassification ? (
                <section className="dataset-classification">
                  <h3>{t.semanticClassification}</h3>
                  <dl>
                    <dt>{t.datasetType}</dt>
                    <dd>{formatDatasetType(datasetClassification.dataset_type, t)}</dd>
                    <dt>{t.scope}</dt>
                    <dd>{formatScope(datasetClassification.scope, t)}</dd>
                    <dt>{t.confidence}</dt>
                    <dd>{formatConfidence(datasetClassification.confidence, language)}</dd>
                    <dt>{t.field}</dt>
                    <dd>{datasetClassification.field_id ? (field?.id === datasetClassification.field_id ? field.name : datasetClassification.field_id) : t.noFieldAssociated}</dd>
                  </dl>
                  <div className="metadata-chips" aria-label={t.missingMetadata}>
                    {(datasetClassification.missing_metadata.length > 0 ? datasetClassification.missing_metadata : [t.noMissingMetadata]).map((item) => (
                      <span key={item}>{formatMissingMetadata(item, t)}</span>
                    ))}
                  </div>
                </section>
              ) : null}
              <section className="dataset-events">
                <h3>{t.datasetHistory}</h3>
                {datasetEvents.length > 0 ? (
                  <div className="event-list">
                    {datasetEvents.map((event) => (
                      <article key={event.event_id} className="event-item">
                        <strong>{formatEventType(event.event_type)}</strong>
                        <span className={`event-status ${event.status}`}>{formatEventStatus(event.status, t)}</span>
                        <small>{formatDate(event.timestamp, language)}</small>
                        <p>{event.message}</p>
                      </article>
                    ))}
                  </div>
                ) : (
                  <p className="muted">{t.noDatasetEvents}</p>
                )}
              </section>
              <form className="dataset-editor" onSubmit={updateSelectedDataset}>
                <label>
                  <span>{t.name}</span>
                  <input value={datasetEditName} disabled={busy} onChange={(event) => setDatasetEditName(event.target.value)} />
                </label>
                <label>
                  <span>{t.source}</span>
                  <input value={datasetEditSourceURI} disabled={busy} onChange={(event) => setDatasetEditSourceURI(event.target.value)} />
                </label>
                <div className="dataset-actions">
                  <button type="submit" disabled={busy}>
                    {t.update}
                  </button>
                  {dataset.archived_at ? (
                    <button type="button" className="secondary-button" disabled={busy} onClick={restoreSelectedDataset}>
                      {t.restore}
                    </button>
                  ) : (
                    <button type="button" className="secondary-button" disabled={busy} onClick={archiveSelectedDataset}>
                      {t.archive}
                    </button>
                  )}
                  {dataset.archived_at ? (
                    <button type="button" className="danger-button" disabled={busy} onClick={deleteSelectedDataset}>
                      {t.hardDelete}
                    </button>
                  ) : null}
                </div>
              </form>
            </>
          ) : (
            <p className="muted">{field ? t.createDatasetHint : t.chooseFieldHint}</p>
          )}
        </aside>

        <section className="panel captures">
          <h2>{t.captures}</h2>
          {captures.map((capture) => (
            <button
              key={capture.id}
              className={capture.id === selectedCaptureID ? 'capture-row selected' : 'capture-row'}
              disabled={busy}
              onClick={() => selectCapture(capture)}
            >
              <div>
                <strong>{capture.capture_key}</strong>
                <span>{formatStatus(capture.validation_status, t)}</span>
              </div>
            </button>
          ))}
          {captures.length === 0 ? <p className="muted">{t.noCaptures}</p> : null}
        </section>
      </section>

      <section className="panel viewer">
        <div className="viewer-header">
          <h2>{t.analysis}</h2>
          <div className="viewer-controls">
            <div className="image-mode-controls" aria-label="Image mode">
              <button className={imageMode === 'rgb' ? 'active' : ''} disabled={!rgbPreview} aria-pressed={imageMode === 'rgb'} onClick={() => selectImageMode('rgb')}>
                {t.rgb}
              </button>
              <button className={imageMode === 'ndvi' ? 'active' : ''} disabled={!ndviPreview} aria-pressed={imageMode === 'ndvi'} onClick={() => selectImageMode('ndvi')}>
                {t.ndvi}
              </button>
            </div>
            <div className="zoom-controls" aria-label={t.zoomLabel}>
              <button disabled={!activePreview || imageZoom <= 100} onClick={() => adjustZoom(-25)}>
                -
              </button>
              <button disabled={!activePreview} onClick={resetZoom}>
                {imageZoom}%
              </button>
              <button className={fitToScreen ? 'active' : ''} disabled={!activePreview} aria-pressed={fitToScreen} onClick={() => setFitToScreen(true)}>
                {t.fit}
              </button>
              <button disabled={!activePreview || imageZoom >= 300} onClick={() => adjustZoom(25)}>
                +
              </button>
            </div>
          </div>
        </div>

        {analysis ? (
          <>
            <div className="analysis-values">
              <Metric label={t.metrics.mean} value={formatMetric(analysis.metrics.mean)} />
              <Metric label={t.metrics.min} value={formatMetric(analysis.metrics.min)} />
              <Metric label={t.metrics.max} value={formatMetric(analysis.metrics.max)} />
              <Metric label={t.metrics.std} value={formatMetric(analysis.metrics.std)} />
              <Metric label={t.metrics.pixels} value={formatPixels(analysis.metrics.valid_pixels, language)} />
            </div>

            <div className="vigor-layout">
              <div className="vigor-values">
                <VigorMetric color="brown" label={t.vigor.nonVegetation} range="< 0.20" value={formatPercent(analysis.metrics.non_vegetation_percent, language)} />
                <VigorMetric color="amber" label={t.vigor.low} range="0.20 - 0.40" value={formatPercent(analysis.metrics.low_vigor_percent, language)} />
                <VigorMetric color="light-green" label={t.vigor.medium} range="0.40 - 0.60" value={formatPercent(analysis.metrics.medium_vigor_percent, language)} />
                <VigorMetric color="green" label={t.vigor.high} range=">= 0.60" value={formatPercent(analysis.metrics.high_vigor_percent, language)} />
              </div>
              <div className="ndvi-legend" aria-label="NDVI">
                <div className="legend-ramp" />
                <div className="legend-labels">
                  <span>-1</span>
                  <span>0</span>
                  <span>1</span>
                </div>
              </div>
            </div>

            <div className="insight-layout">
              <section className="insight-panel">
                <div className="insight-heading">
                  <h2>{t.findings}</h2>
                  <span className={syncStatusClass(analysis.nexus_sync_status)}>{formatSyncStatus(analysis.nexus_sync_status, t)}</span>
                </div>
                {analysis.nexus_sync_error ? <p className="muted">{analysis.nexus_sync_error}</p> : null}
                {analysis.nexus_findings && analysis.nexus_findings.length > 0 ? (
                  <div className="finding-list">
                    {analysis.nexus_findings.map((finding) => (
                      <article key={`${finding.code}-${finding.id ?? ''}`} className="finding-item">
                        <div>
                          <strong>{finding.title}</strong>
                          <span>{finding.code}</span>
                        </div>
                        <small className={`severity ${finding.severity}`}>{finding.severity}</small>
                        <p>{finding.message}</p>
                        {finding.recommendation ? <p className="muted">{finding.recommendation}</p> : null}
                      </article>
                    ))}
                  </div>
                ) : (
                  <p className="muted">{analysis.nexus_sync_status === 'synced' ? t.noFindings : t.unavailable}</p>
                )}
              </section>

              <section className="insight-panel">
                <div className="insight-heading">
                  <h2>{t.assistedInterpretation}</h2>
                  <span className={syncStatusClass(analysis.companion_sync_status)}>{formatSyncStatus(analysis.companion_sync_status, t)}</span>
                </div>
                {analysis.companion_sync_error ? <p className="muted">{analysis.companion_sync_error}</p> : null}
                {analysis.companion_output && Object.keys(analysis.companion_output).length > 0 ? (
                  <div className="assist-output">
                    {assistOutputEntries(analysis.companion_output).map(([key, value]) => (
                      <div key={key}>
                        <strong>{formatAssistKey(key, t)}</strong>
                        <p>{formatAssistValue(value)}</p>
                      </div>
                    ))}
                  </div>
                ) : (
                  <p className="muted">{analysis.companion_sync_status === 'synced' ? t.noInterpretation : t.unavailable}</p>
                )}
              </section>
            </div>
          </>
        ) : null}

        {outputs.length > 0 ? (
          <div className="outputs">
            {outputs.map((output) => (
              <a key={output.id} href={`${apiBase}/v1/assets/${output.id}`} target="_blank">
                {formatOutputKind(output.kind, t)}
              </a>
            ))}
          </div>
        ) : null}

        <div className={fitToScreen ? 'image-stage image-stage-fit' : 'image-stage'}>
          {activePreview ? (
            <img src={activePreview} alt={imageMode === 'rgb' ? t.rgbAlt : t.imageAlt} style={fitToScreen ? undefined : { width: `${imageZoom}%` }} />
          ) : (
            <div className="empty-preview" />
          )}
        </div>
      </section>
    </main>
  );
}

function Metric({ label, value }: { label: string; value: string }) {
  return (
    <div className="metric-card">
      <span>{label}</span>
      <strong>{value}</strong>
    </div>
  );
}

function VigorMetric({ color, label, range, value }: { color: string; label: string; range: string; value: string }) {
  return (
    <div className="vigor-card">
      <span className={`vigor-swatch ${color}`} />
      <div>
        <strong>{value}</strong>
        <span>{label}</span>
        <small>NDVI {range}</small>
      </div>
    </div>
  );
}

function formatMetric(value?: number) {
  return typeof value === 'number' ? value.toFixed(3) : '-';
}

function formatPixels(value: number | undefined, language: Language) {
  return typeof value === 'number' ? new Intl.NumberFormat(language).format(value) : '-';
}

function formatDate(value: string | undefined, language: Language) {
  if (!value) {
    return '-';
  }
  return new Intl.DateTimeFormat(language, {
    dateStyle: 'short',
    timeStyle: 'short',
  }).format(new Date(value));
}

function formatPercent(value: number | undefined, language: Language) {
  if (typeof value !== 'number') {
    return '-';
  }
  return `${new Intl.NumberFormat(language, { maximumFractionDigits: 1, minimumFractionDigits: 1 }).format(value)}%`;
}

function formatStatus(status: string, t: Translation) {
  return (t.statuses as Record<string, string>)[status] ?? status;
}

function formatSyncStatus(status: string | undefined, t: Translation) {
  return (t.syncStatuses as Record<string, string>)[status ?? 'pending'] ?? status ?? 'pending';
}

function formatDatasetType(value: string, t: Translation) {
  return (t.datasetTypes as Record<string, string>)[value] ?? value;
}

function formatScope(value: string, t: Translation) {
  return (t.scopes as Record<string, string>)[value] ?? value;
}

function formatEventStatus(value: string, t: Translation) {
  return (t.eventStatuses as Record<string, string>)[value] ?? value;
}

function formatConfidence(value: number | undefined, language: Language) {
  if (typeof value !== 'number') {
    return '-';
  }
  return `${new Intl.NumberFormat(language, { maximumFractionDigits: 0 }).format(value * 100)}%`;
}

function formatMissingMetadata(value: string, t: Translation) {
  const labels: Record<string, string> = {
    field_id: t.scopes.field,
    lot_id: t.scopes.lot,
    campaign_id: t.scopes.campaign,
    flight_id: t.scopes.flight,
    gps: 'GPS',
    capture_time: 'timestamp',
    drone_metadata: 'drone metadata',
  };
  return labels[value] ?? value;
}

function formatEventType(value: string) {
  return value
    .toLowerCase()
    .split('_')
    .map((part) => part.charAt(0).toUpperCase() + part.slice(1))
    .join(' ');
}

function syncStatusClass(status: string | undefined) {
  return `sync-status ${status ?? 'pending'}`;
}

function assistOutputEntries(output: Record<string, unknown>): [string, unknown][] {
  output = normalizeAssistOutput(output);
  const ordered = ['summary', 'simple_explanation', 'limitations', 'suggested_questions', 'next_steps', 'raw_text'];
  const entries: [string, unknown][] = [];
  for (const key of ordered) {
    if (output[key] !== undefined) {
      if (key === 'raw_text' && output.raw_text === output.summary) {
        continue;
      }
      entries.push([key, output[key]]);
    }
  }
  for (const [key, value] of Object.entries(output)) {
    if (!ordered.includes(key)) {
      entries.push([key, value]);
    }
  }
  return entries;
}

function normalizeAssistOutput(output: Record<string, unknown>): Record<string, unknown> {
  for (const key of ['summary', 'raw_text']) {
    const value = output[key];
    if (typeof value !== 'string') {
      continue;
    }
    const parsed = parseAssistantJSON(value);
    if (parsed) {
      return parsed;
    }
  }
  return output;
}

function parseAssistantJSON(value: string): Record<string, unknown> | null {
  const candidates = [value.trim(), unwrapMarkdownJSON(value), extractJSONObject(value)].filter(Boolean);
  for (const candidate of candidates) {
    try {
      const parsed = JSON.parse(candidate);
      if (parsed && typeof parsed === 'object' && !Array.isArray(parsed)) {
        return parsed as Record<string, unknown>;
      }
    } catch {
      // Try the next representation.
    }
  }
  return null;
}

function unwrapMarkdownJSON(value: string) {
  let text = value.trim();
  if (!text.startsWith('```')) {
    return '';
  }
  text = text.slice(3).trim();
  if (text.toLowerCase().startsWith('json')) {
    text = text.slice(4).trim();
  }
  const fence = text.lastIndexOf('```');
  if (fence >= 0) {
    text = text.slice(0, fence);
  }
  return text.trim();
}

function extractJSONObject(value: string) {
  const start = value.indexOf('{');
  const end = value.lastIndexOf('}');
  return start >= 0 && end > start ? value.slice(start, end + 1).trim() : '';
}

function formatAssistKey(key: string, t: Translation) {
  return (t.assist as Record<string, string>)[key] ?? key;
}

function formatAssistValue(value: unknown): string {
  if (Array.isArray(value)) {
    return value.map(formatAssistValue).join('\n');
  }
  if (value && typeof value === 'object') {
    return JSON.stringify(value);
  }
  if (value === undefined || value === null || value === '') {
    return '-';
  }
  return String(value);
}

function formatOutputKind(kind: string, t: Translation) {
  return (t.outputs as Record<string, string>)[kind] ?? kind;
}

function uploadFileName(file: File) {
  const relative = (file as File & { webkitRelativePath?: string }).webkitRelativePath;
  return relative && relative.trim() ? relative : file.name;
}

function inferUploadDatasetName(files: File[], fallback: string) {
  const first = files[0];
  if (!first) {
    return fallback;
  }
  const relative = (first as File & { webkitRelativePath?: string }).webkitRelativePath;
  const folder = relative?.split('/').filter(Boolean)[0];
  return folder || fallback;
}

createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>,
);
