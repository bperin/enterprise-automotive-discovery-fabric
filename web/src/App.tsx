import React, { useState } from 'react'
import { useDiscovery, DiscoveryResponse, ComparisonReport } from './hooks/use-discovery'
import { 
  Search, ShieldCheck, AlertTriangle, ExternalLink, MapPin, Compass, 
  FileText, Activity, Database, CheckCircle2, Cpu, Zap, Car, Layers, 
  BarChart3, RefreshCw, ChevronRight, Eye, Wrench, Building2, Flame
} from 'lucide-react'

const SAMPLE_QUERIES = [
  { id: 'Q1', category: 'Product Spec', label: 'Towing & Horsepower (Apex Hauler EV)', query: 'What is the towing capacity and horsepower of the 2026 Apex Hauler EV Truck?' },
  { id: 'Q2', category: 'Cross-Brand', label: 'Cross-Brand Comparison (Apex vs Nova)', query: 'Compare the Apex Hauler EV Truck with the Nova Ridge SUV on battery range and towing.' },
  { id: 'Q4', category: 'Live Inventory', label: 'Towing + Live Inventory (Austin)', query: 'Can the 2026 Apex Hauler EV tow my 7,000-pound trailer, and is one available near Austin?' },
  { id: 'Q6', category: 'Conflict Resolution', label: 'Conflicting Spec (Hauler 1500)', query: 'What is the exact payload rating for the Apex Hauler 1500?' },
  { id: 'Q3', category: 'Support RAG', label: 'Support & Software Reset', query: 'How do I reset the digital gauge cluster software on an Apex vehicle?' },
  { id: 'Q7', category: 'Warranty Policy', label: 'Warranty & Commercial Punctures', query: 'Does ApexMotors cover commercial tire punctures under the standard consumer warranty?' },
  { id: 'Q8', category: 'CMS Content', label: 'Newly Published Winter Package', query: 'Show new 2026 winter package updates published in CMS.' },
]

export default function App() {
  const { executeQuery, runBenchmark, loading, benchmarkLoading, error, response, benchmarkReports } = useDiscovery()
  const [queryInput, setQueryInput] = useState('')
  const [postalCode, setPostalCode] = useState('78701')
  const [brand, setBrand] = useState('ApexMotors')
  const [activeTab, setActiveTab] = useState<'results' | 'evidence' | 'trace' | 'benchmark'>('results')

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (!queryInput.trim()) return
    executeQuery({
      query: queryInput,
      brand,
      location: { postal_code: postalCode, radius_miles: 50 },
    })
  }

  const handleSelectQuery = (q: string) => {
    setQueryInput(q)
    executeQuery({
      query: q,
      brand,
      location: { postal_code: postalCode, radius_miles: 50 },
    })
  }

  return (
    <div className="min-h-screen bg-[#07090e] text-slate-100 flex flex-col font-sans selection:bg-cyan-500 selection:text-black">
      {/* Top Navigation Bar */}
      <header className="border-b border-slate-800/80 bg-[#0b0f19]/90 backdrop-blur-xl sticky top-0 z-50 px-6 py-3.5 flex items-center justify-between shadow-2xl">
        <div className="flex items-center space-x-4">
          <div className="relative">
            <div className="absolute -inset-1 bg-gradient-to-r from-cyan-500 to-blue-600 rounded-xl blur opacity-70 animate-pulse"></div>
            <div className="relative bg-slate-900 border border-slate-700/80 p-2.5 rounded-xl text-cyan-400">
              <Cpu className="w-5 h-5" />
            </div>
          </div>
          <div>
            <div className="flex items-center gap-2.5">
              <h1 className="text-sm font-bold tracking-wider text-white uppercase font-mono">
                Gemini Enterprise Automotive Discovery Platform
              </h1>
              <span className="text-[10px] font-mono px-2 py-0.5 rounded-full bg-cyan-500/10 text-cyan-400 border border-cyan-500/30">
                ADK 2.0 Graph v2.4
              </span>
            </div>
            <p className="text-xs text-slate-400 font-mono">Autonomous Multi-Agent Grounded Discovery Cockpit for Major OEM Architecture</p>
          </div>
        </div>

        <div className="flex items-center space-x-4">
          {/* Brand Selector */}
          <div className="flex items-center space-x-2 bg-slate-900/90 px-3 py-1.5 rounded-xl border border-slate-800">
            <Car className="w-4 h-4 text-cyan-400" />
            <select
              value={brand}
              onChange={(e) => setBrand(e.target.value)}
              className="bg-transparent text-xs text-slate-200 font-mono focus:outline-none cursor-pointer"
            >
              <option value="ApexMotors" className="bg-slate-900">ApexMotors (OEM)</option>
              <option value="NovaMotors" className="bg-slate-900">NovaMotors (Affiliate)</option>
            </select>
          </div>

          {/* Postal Code Context */}
          <div className="flex items-center space-x-2 bg-slate-900/90 px-3 py-1.5 rounded-xl border border-slate-800">
            <MapPin className="w-4 h-4 text-cyan-400" />
            <span className="text-xs text-slate-400 font-mono">Geo:</span>
            <input
              type="text"
              value={postalCode}
              onChange={(e) => setPostalCode(e.target.value)}
              className="bg-transparent text-xs w-14 text-white font-mono focus:outline-none text-cyan-300 font-bold"
            />
          </div>

          <button
            onClick={() => {
              setActiveTab('benchmark')
              runBenchmark()
            }}
            className="bg-gradient-to-r from-cyan-600 to-blue-600 hover:from-cyan-500 hover:to-blue-500 text-white text-xs font-mono font-medium px-4 py-2 rounded-xl transition flex items-center gap-2 shadow-lg shadow-cyan-500/25 border border-cyan-400/30"
          >
            <BarChart3 className="w-4 h-4" /> Run 8-Q Benchmark
          </button>
        </div>
      </header>

      {/* Main Container */}
      <main className="flex-1 max-w-7xl w-full mx-auto px-6 py-8 flex flex-col gap-8">
        
        {/* Hero Search & Scenario Hub */}
        <section className="bg-gradient-to-b from-slate-900/90 to-slate-900/40 border border-slate-800/80 rounded-2xl p-6 shadow-2xl backdrop-blur flex flex-col gap-6 relative overflow-hidden">
          <div className="absolute top-0 right-0 w-96 h-96 bg-cyan-500/5 rounded-full blur-3xl pointer-events-none"></div>

          <form onSubmit={handleSubmit} className="flex gap-3 relative z-10">
            <div className="relative flex-1">
              <Search className="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-cyan-400" />
              <input
                type="text"
                value={queryInput}
                onChange={(e) => setQueryInput(e.target.value)}
                placeholder="Ask multi-agent graph: e.g., 'What is the towing capacity and horsepower of the 2026 Apex Hauler EV?'..."
                className="w-full bg-[#0b0f19] border border-slate-700/80 rounded-xl pl-12 pr-4 py-4 text-sm text-white placeholder-slate-500 focus:outline-none focus:border-cyan-500 focus:ring-1 focus:ring-cyan-500 transition shadow-inner font-mono"
              />
            </div>
            <button
              type="submit"
              disabled={loading}
              className="bg-cyan-500 hover:bg-cyan-400 text-black font-bold font-mono px-8 py-4 rounded-xl transition flex items-center gap-2.5 disabled:opacity-50 shadow-xl shadow-cyan-500/20 text-sm tracking-wide cursor-pointer"
            >
              {loading ? (
                <div className="w-5 h-5 border-2 border-black/20 border-t-black rounded-full animate-spin" />
              ) : (
                <Zap className="w-5 h-5 fill-current" />
              )}
              DISCOVER
            </button>
          </form>

          {/* Preset Benchmark Scenarios */}
          <div className="flex flex-col gap-2 relative z-10">
            <div className="flex items-center justify-between">
              <span className="text-[11px] font-mono font-bold text-slate-400 uppercase tracking-widest flex items-center gap-1.5">
                <Layers className="w-3.5 h-3.5 text-cyan-400" /> Enterprise Benchmark Test Prompts
              </span>
              <span className="text-[11px] font-mono text-cyan-400/80">Click any scenario to execute instant ADK graph orchestration</span>
            </div>
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-2.5">
              {SAMPLE_QUERIES.map((sq) => (
                <button
                  key={sq.id}
                  onClick={() => handleSelectQuery(sq.query)}
                  className="text-left bg-slate-950/80 hover:bg-slate-800/80 p-3 rounded-xl border border-slate-800 hover:border-cyan-500/50 transition flex flex-col gap-1 group cursor-pointer"
                >
                  <div className="flex items-center justify-between">
                    <span className="text-[10px] font-mono px-2 py-0.5 rounded bg-cyan-500/10 text-cyan-300 font-bold border border-cyan-500/20">
                      {sq.id} • {sq.category}
                    </span>
                    <ChevronRight className="w-3.5 h-3.5 text-slate-600 group-hover:text-cyan-400 transition transform group-hover:translate-x-0.5" />
                  </div>
                  <span className="text-xs text-slate-300 group-hover:text-white line-clamp-1 font-medium">
                    {sq.label}
                  </span>
                </button>
              ))}
            </div>
          </div>
        </section>

        {/* Error Banner */}
        {error && (
          <div className="bg-red-500/10 border border-red-500/30 rounded-xl p-4 text-xs text-red-300 font-mono flex items-center gap-2">
            <AlertTriangle className="w-4 h-4 text-red-400" />
            <span>Error: {error}</span>
          </div>
        )}

        {/* Benchmark View */}
        {activeTab === 'benchmark' && (
          <div className="bg-slate-900/90 border border-slate-800 rounded-2xl p-6 shadow-2xl flex flex-col gap-6 animate-fadeIn">
            <div className="flex items-center justify-between border-b border-slate-800 pb-4">
              <div>
                <h2 className="text-base font-bold text-white font-mono flex items-center gap-2">
                  <BarChart3 className="w-5 h-5 text-cyan-400" /> 8-Question Benchmark Evaluation Suite
                </h2>
                <p className="text-xs text-slate-400 font-mono mt-0.5">
                  Side-by-side comparison of Legacy Chatbots (Bot A website crawl, Bot B stale RAG, Bot C hallucinating LLM) vs. GCP-Native Unified Discovery Platform.
                </p>
              </div>
              <button
                onClick={runBenchmark}
                disabled={benchmarkLoading}
                className="bg-slate-800 hover:bg-slate-700 text-cyan-400 text-xs font-mono px-4 py-2 rounded-xl border border-slate-700 transition flex items-center gap-2 cursor-pointer"
              >
                {benchmarkLoading ? <RefreshCw className="w-3.5 h-3.5 animate-spin" /> : <RefreshCw className="w-3.5 h-3.5" />}
                Re-Run Evaluation
              </button>
            </div>

            {benchmarkLoading ? (
              <div className="py-20 flex flex-col items-center justify-center gap-4 text-slate-400 font-mono text-xs">
                <div className="w-8 h-8 border-2 border-cyan-500 border-t-transparent rounded-full animate-spin"></div>
                Evaluating all 8 benchmark scenarios across legacy bots and unified platform...
              </div>
            ) : benchmarkReports.length > 0 ? (
              <div className="flex flex-col gap-6">
                {benchmarkReports.map((rep, idx) => (
                  <div key={idx} className="bg-slate-950/80 border border-slate-800/80 rounded-xl p-5 flex flex-col gap-4">
                    <div className="flex items-center justify-between border-b border-slate-900 pb-3">
                      <div className="flex items-center gap-2">
                        <span className="text-xs font-mono font-bold px-2.5 py-1 rounded-lg bg-cyan-500/20 text-cyan-300 border border-cyan-500/30">
                          {rep.case_id}
                        </span>
                        <h3 className="text-sm font-medium text-white">{rep.question}</h3>
                      </div>
                      <span className="text-xs font-mono text-emerald-400 bg-emerald-500/10 px-3 py-1 rounded-full border border-emerald-500/20 flex items-center gap-1 font-bold">
                        <ShieldCheck className="w-3.5 h-3.5" /> Unified Platform: {rep.unified_status}
                      </span>
                    </div>

                    <div className="grid grid-cols-1 md:grid-cols-4 gap-3 text-xs font-mono">
                      {/* Bot A */}
                      <div className="bg-red-950/20 border border-red-500/20 rounded-xl p-3 flex flex-col gap-1.5">
                        <span className="text-red-400 font-bold flex items-center gap-1">
                          <AlertTriangle className="w-3.5 h-3.5" /> Bot A (Website Search)
                        </span>
                        <span className="text-[11px] text-red-300/80">{rep.bot_a_status}</span>
                        <p className="text-slate-400 italic text-[11px] mt-1">"{rep.bot_a_answer}"</p>
                      </div>

                      {/* Bot B */}
                      <div className="bg-red-950/20 border border-red-500/20 rounded-xl p-3 flex flex-col gap-1.5">
                        <span className="text-red-400 font-bold flex items-center gap-1">
                          <AlertTriangle className="w-3.5 h-3.5" /> Bot B (Stale RAG)
                        </span>
                        <span className="text-[11px] text-red-300/80">{rep.bot_b_status}</span>
                        <p className="text-slate-400 italic text-[11px] mt-1">"{rep.bot_b_answer}"</p>
                      </div>

                      {/* Bot C */}
                      <div className="bg-red-950/20 border border-red-500/20 rounded-xl p-3 flex flex-col gap-1.5">
                        <span className="text-red-400 font-bold flex items-center gap-1">
                          <AlertTriangle className="w-3.5 h-3.5" /> Bot C (3rd Party Bot)
                        </span>
                        <span className="text-[11px] text-red-300/80">{rep.bot_c_status}</span>
                        <p className="text-slate-400 italic text-[11px] mt-1">"{rep.bot_c_answer}"</p>
                      </div>

                      {/* Unified */}
                      <div className="bg-cyan-950/20 border border-cyan-500/30 rounded-xl p-3 flex flex-col gap-1.5">
                        <span className="text-cyan-300 font-bold flex items-center gap-1">
                          <ShieldCheck className="w-3.5 h-3.5 text-cyan-400" /> GCP Unified Platform
                        </span>
                        <span className="text-[11px] text-cyan-400">Grounded: {String(rep.unified_grounded)}</span>
                        <p className="text-slate-200 text-[11px] mt-1 font-sans">"{rep.unified_answer}"</p>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            ) : (
              <div className="py-12 text-center text-slate-400 font-mono text-xs">
                Click "Run 8-Q Benchmark" above to execute the side-by-side evaluation suite.
              </div>
            )}
          </div>
        )}

        {/* Results / Query Response */}
        {response ? (
          <div className="flex flex-col gap-6 animate-fadeIn">
            
            {/* Answer & Header */}
            <div className="bg-slate-900/90 border border-slate-800 rounded-2xl p-6 shadow-2xl flex flex-col gap-5 backdrop-blur">
              <div className="flex items-center justify-between border-b border-slate-800 pb-4">
                <div className="flex items-center gap-3">
                  {response.grounded ? (
                    <div className="flex items-center gap-1.5 text-emerald-400 bg-emerald-500/10 px-3.5 py-1.5 rounded-full border border-emerald-500/20 text-xs font-mono font-bold tracking-wide">
                      <ShieldCheck className="w-4 h-4" /> VERIFIED & GROUNDED
                    </div>
                  ) : (
                    <div className="flex items-center gap-1.5 text-amber-400 bg-amber-500/10 px-3.5 py-1.5 rounded-full border border-amber-500/20 text-xs font-mono font-bold tracking-wide">
                      <AlertTriangle className="w-4 h-4" /> UNVERIFIED / QUALIFIED
                    </div>
                  )}
                  <span className="text-xs text-slate-400 font-mono bg-slate-950 px-3 py-1.5 rounded-xl border border-slate-800">
                    Trace: <span className="text-cyan-400">{response.trace_id}</span>
                  </span>
                </div>
                
                {/* Tabs */}
                <div className="flex bg-slate-950 p-1.5 rounded-xl border border-slate-800 font-mono text-xs">
                  <button
                    onClick={() => setActiveTab('results')}
                    className={`px-4 py-2 rounded-lg font-medium transition cursor-pointer ${activeTab === 'results' ? 'bg-cyan-500 text-black font-bold shadow-lg shadow-cyan-500/20' : 'text-slate-400 hover:text-white'}`}
                  >
                    Synthesis & Results ({response.results.length})
                  </button>
                  <button
                    onClick={() => setActiveTab('evidence')}
                    className={`px-4 py-2 rounded-lg font-medium transition cursor-pointer ${activeTab === 'evidence' ? 'bg-cyan-500 text-black font-bold shadow-lg shadow-cyan-500/20' : 'text-slate-400 hover:text-white'}`}
                  >
                    Evidence & Citations ({response.citations.length})
                  </button>
                  <button
                    onClick={() => setActiveTab('trace')}
                    className={`px-4 py-2 rounded-lg font-medium transition cursor-pointer ${activeTab === 'trace' ? 'bg-cyan-500 text-black font-bold shadow-lg shadow-cyan-500/20' : 'text-slate-400 hover:text-white'}`}
                  >
                    ADK Graph Trace
                  </button>
                </div>
              </div>

              {/* Synthesized Answer Card */}
              <div className="bg-[#0b0f19] p-6 rounded-xl border border-slate-800/80 flex flex-col gap-3 shadow-inner">
                <div className="flex items-center justify-between">
                  <span className="text-xs font-mono font-bold text-cyan-400 uppercase tracking-widest flex items-center gap-2">
                    <CheckCircle2 className="w-4 h-4 text-cyan-400" /> Gemini Enterprise Synthesized Answer
                  </span>
                  <span className="text-[10px] font-mono text-slate-500">Model: Gemini 3.5 Flash / Agent Platform</span>
                </div>
                <p className="text-slate-100 text-base leading-relaxed font-sans">{response.answer}</p>
              </div>

              {/* Warnings */}
              {response.warnings && response.warnings.length > 0 && (
                <div className="bg-amber-500/10 border border-amber-500/20 rounded-xl p-4 flex flex-col gap-1.5 text-xs text-amber-300 font-mono">
                  <span className="font-bold flex items-center gap-1.5">
                    <AlertTriangle className="w-4 h-4 text-amber-400" /> Compliance & Grounding Warnings:
                  </span>
                  {response.warnings.map((w, idx) => (
                    <span key={idx} className="text-amber-200/80">• [{w.code}] {w.message}</span>
                  ))}
                </div>
              )}

              {/* Suggested Actions */}
              {response.actions && response.actions.length > 0 && (
                <div className="flex flex-wrap items-center gap-3 pt-1">
                  <span className="text-xs font-mono text-slate-400">Recommended OEM Actions:</span>
                  {response.actions.map((act, idx) => (
                    <a
                      key={idx}
                      href={act.url}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="text-xs font-mono bg-cyan-500/10 hover:bg-cyan-500/20 text-cyan-300 px-4 py-2.5 rounded-xl border border-cyan-500/30 transition flex items-center gap-2 font-medium shadow-md shadow-cyan-500/10"
                    >
                      {act.label} <ExternalLink className="w-3.5 h-3.5 text-cyan-400" />
                    </a>
                  ))}
                </div>
              )}
            </div>

            {/* Tab Contents */}
            {activeTab === 'results' && (
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                {response.results.map((res, idx) => (
                  <div key={idx} className="bg-slate-900/90 border border-slate-800/90 rounded-2xl p-6 flex flex-col justify-between gap-5 shadow-xl hover:border-cyan-500/40 transition">
                    <div className="flex flex-col gap-3">
                      <div className="flex items-center justify-between">
                        <span className="text-xs font-mono font-bold px-3 py-1 rounded-full bg-slate-950 text-cyan-400 border border-slate-800">
                          {res.entity_type || 'Vehicle / Product'}
                        </span>
                        <span className="text-xs font-mono text-slate-400 bg-slate-950 px-2.5 py-1 rounded-lg border border-slate-800">
                          Relevance: {(res.score * 100).toFixed(0)}%
                        </span>
                      </div>
                      <h4 className="text-lg font-bold text-white font-sans">{res.title}</h4>
                      <p className="text-sm text-slate-300 leading-relaxed font-sans">{res.snippet}</p>
                    </div>

                    <div className="border-t border-slate-800/80 pt-4 flex items-center justify-between text-xs font-mono text-slate-400">
                      <span className="flex items-center gap-1.5 text-cyan-300">
                        <Database className="w-4 h-4 text-cyan-400" /> {res.source_ref.source_id} <span className="text-slate-500">({res.source_ref.source_type})</span>
                      </span>
                      <a
                        href={res.source_ref.source_uri}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="text-cyan-400 hover:underline flex items-center gap-1 font-bold"
                      >
                        Inspect URI <ExternalLink className="w-3.5 h-3.5" />
                      </a>
                    </div>
                  </div>
                ))}
              </div>
            )}

            {activeTab === 'evidence' && (
              <div className="bg-slate-900/90 border border-slate-800 rounded-2xl p-6 shadow-2xl flex flex-col gap-4">
                <h3 className="text-sm font-mono font-bold text-white uppercase tracking-wider flex items-center gap-2">
                  <FileText className="w-4 h-4 text-cyan-400" /> Evidence Drawer & Source Citations
                </h3>
                <div className="flex flex-col gap-3">
                  {response.citations.map((cit, idx) => (
                    <div key={idx} className="bg-[#0b0f19] p-4 rounded-xl border border-slate-800 flex flex-col gap-2 font-mono">
                      <div className="flex items-center justify-between">
                        <span className="text-xs font-bold text-cyan-400">{cit.title}</span>
                        <span className="text-xs text-slate-500 bg-slate-900 px-2.5 py-1 rounded border border-slate-800">{cit.source_id}</span>
                      </div>
                      <p className="text-xs text-slate-300 italic font-sans bg-slate-900/50 p-3 rounded border border-slate-800/80">"{cit.excerpt}"</p>
                      <a href={cit.uri} target="_blank" rel="noopener noreferrer" className="text-xs text-cyan-400 hover:underline flex items-center gap-1 font-bold">
                        {cit.uri} <ExternalLink className="w-3.5 h-3.5" />
                      </a>
                    </div>
                  ))}
                </div>
              </div>
            )}

            {activeTab === 'trace' && (
              <div className="bg-slate-900/90 border border-slate-800 rounded-2xl p-6 shadow-2xl flex flex-col gap-5">
                <h3 className="text-sm font-mono font-bold text-white uppercase tracking-wider flex items-center gap-2">
                  <Activity className="w-4 h-4 text-cyan-400" /> ADK 2.0 Multi-Agent Graph Workflow Execution Trace
                </h3>
                <div className="flex flex-col gap-3 font-mono text-xs">
                  <div className="bg-[#0b0f19] p-4 rounded-xl border border-slate-800 flex items-center justify-between">
                    <div className="flex items-center gap-3">
                      <div className="w-2 h-2 rounded-full bg-emerald-400 animate-ping"></div>
                      <span className="text-emerald-400 font-bold">[1] RetrievalPlannerAgent</span>
                    </div>
                    <span className="text-slate-400">Classified query intent & planned specialist fan-out</span>
                  </div>
                  <div className="bg-[#0b0f19] p-4 rounded-xl border border-slate-800 flex items-center justify-between">
                    <div className="flex items-center gap-3">
                      <div className="w-2 h-2 rounded-full bg-emerald-400"></div>
                      <span className="text-emerald-400 font-bold">[2] Parallel Specialist Execution</span>
                    </div>
                    <span className="text-slate-400">Queried Agent Search, Vector RAG & Live Inventory SQL</span>
                  </div>
                  <div className="bg-[#0b0f19] p-4 rounded-xl border border-slate-800 flex items-center justify-between">
                    <div className="flex items-center gap-3">
                      <div className="w-2 h-2 rounded-full bg-emerald-400"></div>
                      <span className="text-emerald-400 font-bold">[3] AuthorityResolverAgent</span>
                    </div>
                    <span className="text-slate-400">Enforced Spec DB hierarchy over unverified web copy</span>
                  </div>
                  <div className="bg-[#0b0f19] p-4 rounded-xl border border-slate-800 flex items-center justify-between">
                    <div className="flex items-center gap-3">
                      <div className="w-2 h-2 rounded-full bg-emerald-400"></div>
                      <span className="text-emerald-400 font-bold">[4] GroundedSynthesizerAgent</span>
                    </div>
                    <span className="text-slate-400">Synthesized grounded response with verified citations</span>
                  </div>
                  <div className="bg-[#0b0f19] p-4 rounded-xl border border-slate-800 flex items-center justify-between">
                    <div className="flex items-center gap-3">
                      <div className="w-2 h-2 rounded-full bg-emerald-400"></div>
                      <span className="text-emerald-400 font-bold">[5] GroundingGate & Policy Check</span>
                    </div>
                    <span className="text-cyan-400 font-bold">Verdict: PASS (Trace ID: {response.trace_id})</span>
                  </div>
                </div>
              </div>
            )}

          </div>
        ) : (
          <div className="bg-slate-900/40 border border-slate-800/80 rounded-2xl p-16 text-center flex flex-col items-center justify-center gap-5 backdrop-blur">
            <div className="relative">
              <div className="absolute -inset-2 bg-cyan-500/20 rounded-full blur-xl animate-pulse"></div>
              <div className="relative bg-slate-900 border border-slate-700 p-5 rounded-2xl text-cyan-400">
                <Compass className="w-10 h-10 animate-spin-slow" />
              </div>
            </div>
            <div className="flex flex-col gap-1.5 max-w-lg">
              <h3 className="text-lg font-bold text-white font-mono">Gemini Enterprise Discovery Cockpit</h3>
              <p className="text-xs text-slate-400 font-mono leading-relaxed">
                Select an enterprise benchmark prompt above or enter a custom query to initiate the ADK 2.0 multi-agent graph orchestration pipeline for major automotive manufacturer discovery.
              </p>
            </div>
          </div>
        )}

      </main>
    </div>
  )
}
