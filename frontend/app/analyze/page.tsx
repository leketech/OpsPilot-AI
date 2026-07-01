"use client";

import { useState } from "react";
import IncidentForm from "@/components/IncidentForm";
import PipelineView from "@/components/PipelineView";
import ResultView from "@/components/ResultView";
import { analyzeIncident } from "@/lib/api";
import type { AnalysisResult, IncidentInput } from "@/lib/types";

// Each step duration roughly mirrors how long that LLM call takes.
// The animation runs in parallel with the real API call — whichever finishes
// last determines when results appear.
const PIPELINE_STEPS = [
  {
    label: "Gathering Evidence",
    sub: "kubernetes · prometheus · github · argocd · memory",
    duration: 4000,
  },
  {
    label: "Investigating",
    sub: "Investigator Agent",
    duration: 3000,
  },
  {
    label: "Planning Remediation",
    sub: "Planner Agent",
    duration: 3000,
  },
  {
    label: "Generating Runbook",
    sub: "Executor Agent",
    duration: 2500,
  },
  {
    label: "Safety Review",
    sub: "Reviewer Agent",
    duration: 2500,
  },
];

type UIState = "idle" | "loading" | "result" | "error";

const sleep = (ms: number) => new Promise<void>((r) => setTimeout(r, ms));

export default function AnalyzePage() {
  const [uiState, setUiState]       = useState<UIState>("idle");
  const [activeStep, setActiveStep] = useState(-1);
  const [result, setResult]         = useState<AnalysisResult | null>(null);
  const [error, setError]           = useState<string | null>(null);

  async function handleSubmit(incident: IncidentInput) {
    setUiState("loading");
    setResult(null);
    setError(null);
    setActiveStep(0);

    // Kick off the API call immediately — don't await it yet.
    let apiResult: AnalysisResult | null = null;
    let apiError: string | null = null;
    const apiPromise = analyzeIncident(incident)
      .then((r) => { apiResult = r; })
      .catch((e: Error) => { apiError = e.message; });

    // Animate through all pipeline steps in parallel with the API call.
    for (let i = 0; i < PIPELINE_STEPS.length; i++) {
      setActiveStep(i);
      await sleep(PIPELINE_STEPS[i].duration);
    }
    setActiveStep(PIPELINE_STEPS.length); // mark all complete

    // Wait for the API if it hasn't finished yet.
    await apiPromise;

    if (apiResult) {
      setResult(apiResult);
      setUiState("result");
    } else {
      setError(apiError ?? "Unknown error occurred.");
      setUiState("error");
    }
  }

  function reset() {
    setUiState("idle");
    setResult(null);
    setError(null);
    setActiveStep(-1);
  }

  return (
    <div className="h-[calc(100vh-49px)] grid grid-cols-1 lg:grid-cols-[420px_1fr] divide-x divide-slate-800 overflow-hidden">
      {/* ── Left: Form ──────────────────────────────────────────── */}
      <div className="overflow-y-auto p-6">
        <IncidentForm onSubmit={handleSubmit} disabled={uiState === "loading"} />

        {(uiState === "result" || uiState === "error") && (
          <button
            onClick={reset}
            className="mt-4 w-full text-sm text-slate-400 hover:text-white border border-slate-700 hover:border-slate-500 py-2.5 rounded-lg transition"
          >
            ← Analyze another incident
          </button>
        )}
      </div>

      {/* ── Right: Pipeline / Results / Error ───────────────────── */}
      <div className="overflow-y-auto p-6">
        {uiState === "idle" && (
          <div className="h-full flex flex-col items-center justify-center text-center gap-3">
            <div className="w-12 h-12 rounded-full border border-slate-700 flex items-center justify-center text-slate-600 text-xl">
              ⬡
            </div>
            <p className="text-slate-500 text-sm max-w-xs">
              Fill in the incident details on the left and click{" "}
              <span className="text-slate-300">Analyze Incident</span> to start
              the agent pipeline.
            </p>
            <p className="text-slate-600 text-xs">
              Or use <span className="text-slate-400">Load Sample</span> to run a
              pre-built demo.
            </p>
          </div>
        )}

        {uiState === "loading" && (
          <PipelineView steps={PIPELINE_STEPS} activeStep={activeStep} />
        )}

        {uiState === "result" && result && <ResultView result={result} />}

        {uiState === "error" && (
          <div className="border border-red-900 bg-red-950/20 rounded-lg p-5 space-y-2">
            <p className="text-red-400 font-semibold text-sm">Analysis failed</p>
            <p className="text-red-300/70 text-xs font-mono">{error}</p>
            <p className="text-slate-500 text-xs mt-1">
              Make sure the Go backend is running on port 8080 and{" "}
              <code className="text-slate-400">QWEN_API_KEY</code> is set in{" "}
              <code className="text-slate-400">.env</code>.
            </p>
          </div>
        )}
      </div>
    </div>
  );
}
