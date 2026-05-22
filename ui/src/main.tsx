import React, { useEffect, useMemo, useState } from 'react';
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
    scan: 'Escanear',
    processing: 'Procesando',
    dataset: 'Dataset',
    name: 'Nombre',
    status: 'Estado',
    source: 'Origen',
    createDatasetHint: 'Crea un dataset desde la ruta local de muestra.',
    captures: 'Capturas',
    noCaptures: 'Todavia no hay capturas escaneadas.',
    analysis: 'Analisis',
    rgb: 'RGB',
    ndvi: 'NDVI',
    zoomLabel: 'Zoom de imagen',
    fit: 'Ajustar',
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
  },
  en: {
    tagline: 'Multispectral drone vision pipeline',
    languageLabel: 'Language',
    sampleDatasetName: 'Sample DJI M3M',
    scan: 'Scan',
    processing: 'Processing',
    dataset: 'Dataset',
    name: 'Name',
    status: 'Status',
    source: 'Source',
    createDatasetHint: 'Create a dataset from the local sample path.',
    captures: 'Captures',
    noCaptures: 'No captures scanned yet.',
    analysis: 'Analysis',
    rgb: 'RGB',
    ndvi: 'NDVI',
    zoomLabel: 'Image zoom',
    fit: 'Fit',
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
  },
} as const;

type Translation = (typeof translations)[Language];

type Dataset = {
  id: string;
  name: string;
  source_uri: string;
  status: string;
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
  kind: string;
  status: string;
  metrics: Record<string, number>;
  outputs: OutputAsset[];
};

type OutputAsset = {
  id: string;
  kind: string;
  content_type: string;
  byte_size: number;
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
  const [language, setLanguage] = useState<Language>(initialLanguage);
  const [sourceURI, setSourceURI] = useState(defaultSourceURI);
  const [dataset, setDataset] = useState<Dataset | null>(null);
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

  const ndviPreview = useMemo(() => {
    const output = analysis?.outputs.find((item) => item.kind === 'preview_png');
    return output ? `${apiBase}/v1/assets/${output.id}` : '';
  }, [analysis]);

  const rgbPreview = selectedCaptureID ? `${apiBase}/v1/captures/${selectedCaptureID}/assets/rgb` : '';
  const activePreview = imageMode === 'rgb' ? rgbPreview : ndviPreview;
  const outputs = useMemo(() => analysis?.outputs ?? [], [analysis]);

  async function createAndScan() {
    setBusy(true);
    setError(null);
    setAnalysis(null);
    setSelectedCaptureID(null);
    setImageMode('rgb');
    setImageZoom(100);
    setFitToScreen(true);
    try {
      const created = await request<Dataset>('/v1/datasets', {
        method: 'POST',
        body: JSON.stringify({ name: t.sampleDatasetName, source_uri: sourceURI }),
      });
      setDataset(created);
      const scan = await request<{ captures: Capture[] }>(`/v1/datasets/${created.id}/scan`, {
        method: 'POST',
      });
      setCaptures(scan.captures);
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
            <button disabled={busy} onClick={createAndScan}>
              {busy ? t.processing : t.scan}
            </button>
          </div>
        </div>
      </section>

      {error ? <pre className="error">{error}</pre> : null}

      <section className="summary-layout">
        <aside className="panel">
          <h2>{t.dataset}</h2>
          {dataset ? (
            <dl>
              <dt>{t.name}</dt>
              <dd>{dataset.name}</dd>
              <dt>{t.status}</dt>
              <dd>{formatStatus(dataset.status, t)}</dd>
              <dt>{t.source}</dt>
              <dd>{dataset.source_uri}</dd>
            </dl>
          ) : (
            <p className="muted">{t.createDatasetHint}</p>
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

function formatPercent(value: number | undefined, language: Language) {
  if (typeof value !== 'number') {
    return '-';
  }
  return `${new Intl.NumberFormat(language, { maximumFractionDigits: 1, minimumFractionDigits: 1 }).format(value)}%`;
}

function formatStatus(status: string, t: Translation) {
  return (t.statuses as Record<string, string>)[status] ?? status;
}

function formatOutputKind(kind: string, t: Translation) {
  return (t.outputs as Record<string, string>)[kind] ?? kind;
}

createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>,
);
