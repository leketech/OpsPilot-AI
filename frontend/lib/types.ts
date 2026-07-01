export type Severity = "critical" | "high" | "medium" | "low";
export type RiskLevel = "low" | "medium" | "high" | "critical";

export interface IncidentInput {
  id: string;
  title: string;
  service: string;
  severity: Severity;
  namespace: string;
  cluster: string;
  description: string;
  evidence?: {
    logs?: string;
    metrics?: string;
    events?: string;
  };
}

export interface ExecutionStep {
  order: number;
  type: string;
  command: string;
  purpose: string;
  risk: RiskLevel;
  rollback: string;
}

export interface AnalysisResult {
  rootCause: string;
  confidence: number;
  recommendedAction: string;
  risk: string;
  similarIncidents: number;
  investigation: {
    summary: string;
    keyFindings: string;
    dataGaps: string;
  };
  executionPlan: {
    steps: ExecutionStep[];
    requires: string;
  };
  review: {
    approved: boolean;
    overallRisk: string;
    concerns: string;
    modifications: string;
    finalVerdict: string;
  };
}
