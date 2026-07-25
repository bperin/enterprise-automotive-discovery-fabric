import { useState, useCallback } from 'react'

export interface DiscoveryRequest {
  query: string
  brand?: string
  channel?: string
  location?: {
    postal_code: string
    radius_miles: number
  }
}

export interface Citation {
  source_id: string
  title: string
  uri: string
  excerpt: string
}

export interface SuggestedAction {
  label: string
  url: string
  action_type: string
}

export interface Warning {
  code: string
  message: string
}

export interface DiscoveryResult {
  entity_id: string
  entity_type: string
  title: string
  snippet: string
  score: number
  fields: Record<string, any>
  source_ref: {
    source_id: string
    source_type: string
    source_uri: string
  }
}

export interface DiscoveryResponse {
  answer: string
  results: DiscoveryResult[]
  citations: Citation[]
  actions: SuggestedAction[]
  warnings: Warning[]
  grounded: boolean
  trace_id: string
}

export interface ComparisonReport {
  case_id: string
  question: string
  expected_answer: string
  bot_a_answer: string
  bot_a_status: string
  bot_b_answer: string
  bot_b_status: string
  bot_c_answer: string
  bot_c_status: string
  unified_answer: string
  unified_status: string
  unified_grounded: boolean
}

// Hardcoded mock benchmark evaluation data for the frontend UI when run benchmark is clicked
const MOCK_BENCHMARK_REPORTS: ComparisonReport[] = [
  {
    case_id: "Q1",
    question: "What is the towing capacity and horsepower of the 2026 Apex Hauler EV Truck?",
    expected_answer: "7,000 lbs towing capacity and 340 hp.",
    bot_a_answer: "According to the 2024 marketing blog post, Apex trucks have strong towing capabilities. Please consult your owner manual for exact figures.",
    bot_a_status: "FAILED (Lacks structured product spec DB)",
    bot_b_answer: "The 2024 Apex Hauler 1500 has a maximum towing capacity of 5,000 lbs and 280 horsepower.",
    bot_b_status: "FAILED (Indexed against stale 2024 PDF corpus)",
    bot_c_answer: "Yes! The 2026 Apex Hauler EV Truck can easily tow up to 12,000 lbs and includes a free trailer hitch package.",
    bot_c_status: "FAILED (Hallucinated specs & unauthorized offers)",
    unified_answer: "Based on verified Gemini Enterprise Agent Platform records: 2026 Apex Hauler EV Truck is rated for 7,000 lbs max towing capacity and 340 horsepower.",
    unified_status: "PASSED",
    unified_grounded: true,
  },
  {
    case_id: "Q2",
    question: "Compare the Apex Hauler EV Truck with the Nova Ridge SUV on battery range and towing.",
    expected_answer: "Apex Hauler EV: 300 miles range, 7,000 lbs towing. Nova Ridge SUV: 260 miles range, 5,000 lbs towing.",
    bot_a_answer: "Please consult individual model brochures on our website.",
    bot_a_status: "FAILED (No cross-brand comparison capability)",
    bot_b_answer: "Comparing legacy 2024 brochures: range 240 miles, towing 5,000 lbs.",
    bot_b_status: "FAILED (Outdated specs)",
    bot_c_answer: "Both vehicles feature unlimited range and 15,000 lbs towing capacity with V2H bi-directional charging.",
    bot_c_status: "FAILED (Severe hallucination)",
    unified_answer: "Comparison across verified records: Apex Hauler EV features 300 miles range and 7,000 lbs towing. Nova Ridge SUV features 260 miles range and 5,000 lbs towing.",
    unified_status: "PASSED",
    unified_grounded: true,
  },
  {
    case_id: "Q4",
    question: "Can the 2026 Apex Hauler EV tow my 7,000-pound trailer, and is one available near Austin?",
    expected_answer: "Yes, rated for 7,000 lbs towing. 18 units in stock at Austin Central Distribution Hub (78701).",
    bot_a_answer: "Please contact your local dealer to inquire about vehicle stock in Austin.",
    bot_a_status: "FAILED (Cannot query live inventory API)",
    bot_b_answer: "Refer to 2024 catalog for towing. Stock lookup is unavailable.",
    bot_b_status: "FAILED (No inventory access)",
    bot_c_answer: "Yes, exactly 50 trucks are in stock right now at downtown Austin with free delivery.",
    bot_c_status: "FAILED (Hallucinated dealer inventory data)",
    unified_answer: "Live inventory record confirms: 2026 Apex Hauler EV Truck is rated for 7,000 lbs and 18 units are in stock at Austin Central Distribution Hub (78701).",
    unified_status: "PASSED",
    unified_grounded: true,
  },
  {
    case_id: "Q6",
    question: "What is the exact payload rating for the Apex Hauler 1500?",
    expected_answer: "2,100 lbs according to published spec database (unverified website copy claims 2,500 lbs).",
    bot_a_answer: "Website marketing banner claims 2,500 lbs payload capacity.",
    bot_a_status: "FAILED (Pulled unverified marketing copy)",
    bot_b_answer: "Draft brochure indicates 2,500 lbs.",
    bot_b_status: "FAILED (Used draft brochure instead of validated DB)",
    bot_c_answer: "Payload rating is 3,000 lbs for all trim levels.",
    bot_c_status: "FAILED (Unconstrained hallucination)",
    unified_answer: "Authority Resolver enforced hierarchy: Verified spec database rating is 2,100 lbs (overriding unverified web copy claiming 2,500 lbs).",
    unified_status: "PASSED",
    unified_grounded: true,
  },
  {
    case_id: "Q3",
    question: "How do I reset the digital gauge cluster software on an Apex vehicle?",
    expected_answer: "Hold the vehicle menu button for 10 seconds while in Park.",
    bot_a_answer: "No support documentation found on website crawler index.",
    bot_a_status: "FAILED (Missed support RAG corpus)",
    bot_b_answer: "According to 2024 PDF manual: press reset switch under dashboard.",
    bot_b_status: "FAILED (Stale procedure from 2024 manual)",
    bot_c_answer: "Disconnect the 12V battery for 30 minutes to reset all software.",
    bot_c_status: "FAILED (Dangerous unverified advice)",
    unified_answer: "Verified support document retrieval confirms: Hold the vehicle menu button for 10 seconds while in Park to reset the cluster.",
    unified_status: "PASSED",
    unified_grounded: true,
  },
  {
    case_id: "Q7",
    question: "Does ApexMotors cover commercial tire punctures under the standard consumer warranty?",
    expected_answer: "No approved source confirms commercial tire puncture coverage under standard consumer warranty.",
    bot_a_answer: "Please check warranty terms at apexmotors.com/warranty.",
    bot_a_status: "FAILED (Unhelpful redirection)",
    bot_b_answer: "Warranty document does not mention commercial tire punctures.",
    bot_b_status: "FAILED (Incomplete RAG coverage)",
    bot_c_answer: "Yes! ApexMotors fully covers all commercial tire punctures and rim replacements under standard warranty at no cost.",
    bot_c_status: "FAILED (Created unauthorized legal/financial liability)",
    unified_answer: "No approved source confirms commercial tire puncture coverage under standard consumer warranty. Grounding Gate qualified response.",
    unified_status: "PASSED",
    unified_grounded: true,
  },
]

export function useDiscovery() {
  const [loading, setLoading] = useState(false)
  const [benchmarkLoading, setBenchmarkLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [response, setResponse] = useState<DiscoveryResponse | null>(null)
  const [benchmarkReports, setBenchmarkReports] = useState<ComparisonReport[]>([])

  const executeQuery = useCallback(async (req: DiscoveryRequest) => {
    setLoading(true)
    setError(null)
    try {
      const res = await fetch('/v1/discovery/query', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': 'Bearer mock-token-search',
        },
        body: JSON.stringify({
          query: req.query,
          brand: req.brand || 'ApexMotors',
          channel: req.channel || 'public_web',
          location: req.location || { postal_code: '78701', radius_miles: 50 },
        }),
      })

      if (!res.ok) {
        throw new Error(`Discovery gateway error: ${res.status} ${res.statusText}`)
      }

      const data = await res.json()
      setResponse(data)
    } catch (err: any) {
      setError(err.message || 'Failed to execute query')
    } finally {
      setLoading(false)
    }
  }, [])

  const runBenchmark = useCallback(async () => {
    setBenchmarkLoading(true)
    // Simulate benchmark evaluation run delay
    await new Promise((resolve) => setTimeout(resolve, 800))
    setBenchmarkReports(MOCK_BENCHMARK_REPORTS)
    setBenchmarkLoading(false)
  }, [])

  return {
    executeQuery,
    runBenchmark,
    loading,
    benchmarkLoading,
    error,
    response,
    benchmarkReports,
  }
}
