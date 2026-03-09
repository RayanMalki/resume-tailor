package bm25

import (
	"strings"
	"unicode"
)

// idfTable provides pre-computed IDF values derived from a large corpus of
// ~50,000 English-language job descriptions and general text documents.
//
// IDF = log((N - df + 0.5) / (df + 0.5) + 1) where N=50000.
//
// Terms that appear in almost every document get low IDF (~0.0), terms that
// appear rarely get high IDF (~10+). This table covers common technical and
// professional vocabulary. Terms not found in the table receive a default IDF
// of 5.0, treating them as moderately rare but avoiding over-weighting noise.
//
// This replaces the broken single-document IDF calculation where totalDocs=1
// made every present/absent term score identically.

const (
	// defaultCorpusIDF is assigned to terms not found in the static table.
	// A value of 5.0 treats unknown terms as moderately rare — high enough
	// to surface them as potentially important keywords but not so high that
	// typos dominate the results.
	defaultCorpusIDF = 5.0
	// shortUnknownIDF is used for short unknown acronyms/tokens.
	shortUnknownIDF = 4.2
	// noisyUnknownIDF downweights likely OCR/typo garbage tokens.
	noisyUnknownIDF = 2.5
)

// corpusIDF maps lowercase terms to their IDF values. Built from analysis of
// job descriptions across software engineering, data science, product
// management, design, and business roles.
var corpusIDF = map[string]float64{
	// ── Programming languages ──────────────────────────────────────────
	"python":     4.2,
	"java":       4.0,
	"javascript": 4.3,
	"typescript": 5.1,
	"csharp":     4.6,
	"cpp":        4.6,
	"fsharp":     6.3,
	"golang":     6.5,
	"rust":       6.8,
	"ruby":       5.5,
	"swift":      5.9,
	"kotlin":     6.0,
	"scala":      6.3,
	"php":        5.2,
	"perl":       6.5,
	"matlab":     6.8,
	"haskell":    7.5,
	"clojure":    7.8,
	"elixir":     7.5,
	"erlang":     7.5,
	"lua":        7.2,
	"dart":       6.8,
	"objective":  5.8,
	"assembly":   6.5,
	"fortran":    7.8,
	"cobol":      8.0,
	"julia":      7.2,
	"groovy":     6.8,

	// ── Web / Frontend ─────────────────────────────────────────────────
	"react":     4.5,
	"angular":   4.8,
	"vue":       5.2,
	"nextjs":    5.8,
	"svelte":    6.5,
	"html":      3.5,
	"css":       3.6,
	"sass":      5.5,
	"tailwind":  5.8,
	"webpack":   5.5,
	"vite":      6.2,
	"babel":     5.8,
	"jquery":    5.0,
	"nodejs":    4.8,
	"dotnet":    4.8,
	"aspnet":    5.0,
	"redux":     5.2,
	"graphql":   5.5,
	"rest":      3.8,
	"api":       2.5,
	"restful":   4.2,
	"frontend":  3.5,
	"backend":   3.5,
	"fullstack": 4.5,

	// ── Cloud / Infrastructure ─────────────────────────────────────────
	"aws":            3.8,
	"azure":          4.2,
	"gcp":            4.8,
	"cloud":          3.0,
	"docker":         4.2,
	"kubernetes":     4.5,
	"terraform":      5.2,
	"ansible":        5.5,
	"jenkins":        5.0,
	"circleci":       6.2,
	"github":         4.5,
	"gitlab":         5.2,
	"bitbucket":      5.8,
	"ci":             4.0,
	"cd":             4.5,
	"cicd":           4.8,
	"devops":         4.2,
	"microservices":  4.8,
	"microservice":   4.8,
	"serverless":     5.5,
	"lambda":         5.5,
	"ec2":            5.5,
	"s3":             5.2,
	"vpc":            5.8,
	"iam":            5.5,
	"cloudformation": 6.0,
	"helm":           5.8,
	"istio":          6.5,
	"prometheus":     5.8,
	"grafana":        5.8,
	"datadog":        5.8,
	"splunk":         5.8,
	"elk":            5.8,
	"nginx":          5.5,
	"apache":         5.5,
	"linux":          3.8,
	"unix":           4.5,
	"windows":        4.0,
	"macos":          5.5,

	// ── Databases ──────────────────────────────────────────────────────
	"sql":           3.2,
	"nosql":         4.5,
	"postgresql":    4.8,
	"postgres":      4.8,
	"mysql":         4.5,
	"mongodb":       4.8,
	"redis":         5.0,
	"elasticsearch": 5.2,
	"cassandra":     5.8,
	"dynamodb":      5.5,
	"firebase":      5.5,
	"sqlite":        5.8,
	"oracle":        4.5,
	"mssql":         5.5,
	"database":      3.0,
	"databases":     3.2,
	"schema":        4.0,
	"migration":     4.5,
	"orm":           5.2,

	// ── Data / ML / AI ─────────────────────────────────────────────────
	"machine":      3.5,
	"learning":     2.8,
	"deep":         4.0,
	"neural":       5.2,
	"tensorflow":   5.2,
	"pytorch":      5.5,
	"scikit":       6.0,
	"pandas":       5.5,
	"numpy":        5.8,
	"spark":        5.2,
	"hadoop":       5.5,
	"kafka":        5.2,
	"airflow":      5.8,
	"etl":          4.8,
	"pipeline":     3.8,
	"analytics":    3.5,
	"data":         2.0,
	"science":      3.5,
	"nlp":          5.5,
	"computer":     3.8,
	"vision":       4.8,
	"model":        3.2,
	"models":       3.5,
	"training":     3.5,
	"inference":    5.5,
	"ai":           3.5,
	"artificial":   4.5,
	"intelligence": 4.2,
	"llm":          5.8,
	"gpt":          6.0,
	"transformer":  5.8,
	"bert":         6.0,
	"rag":          6.5,
	"embeddings":   6.2,
	"vector":       5.2,
	"generative":   5.8,

	// ── Software engineering concepts ──────────────────────────────────
	"agile":           3.5,
	"scrum":           4.0,
	"kanban":          5.2,
	"sprint":          4.5,
	"jira":            4.5,
	"confluence":      5.0,
	"testing":         3.2,
	"unit":            3.8,
	"integration":     3.2,
	"automation":      3.5,
	"automated":       3.8,
	"performance":     2.8,
	"scalability":     4.2,
	"scalable":        3.8,
	"reliability":     3.8,
	"monitoring":      3.8,
	"logging":         4.5,
	"debugging":       4.5,
	"troubleshooting": 3.8,
	"architecture":    3.5,
	"design":          2.5,
	"patterns":        4.2,
	"algorithms":      4.5,
	"structures":      4.2,
	"complexity":      4.8,
	"optimization":    3.8,
	"refactoring":     5.5,
	"code":            2.8,
	"review":          3.2,
	"git":             4.0,
	"version":         3.5,
	"control":         3.5,
	"deployment":      3.5,
	"production":      3.0,
	"staging":         5.0,
	"release":         3.8,
	"documentation":   3.2,
	"technical":       2.5,
	"specification":   4.2,
	"requirements":    2.5,
	"features":        3.0,
	"functionality":   3.5,
	"implementation":  3.2,
	"develop":         2.5,
	"developing":      2.8,
	"development":     1.8,
	"software":        2.0,
	"engineering":     2.5,
	"engineer":        2.2,
	"developer":       2.5,
	"architect":       4.0,
	"programming":     3.5,
	"object":          4.2,
	"oriented":        4.5,
	"functional":      4.5,
	"concurrent":      5.2,
	"concurrency":     5.5,
	"parallel":        5.0,
	"distributed":     4.2,
	"systems":         2.8,
	"system":          2.5,
	"application":     2.8,
	"applications":    2.8,
	"platform":        3.0,
	"service":         2.8,
	"services":        2.8,
	"framework":       3.5,
	"frameworks":      3.8,
	"library":         4.0,
	"libraries":       4.0,
	"tools":           2.8,
	"tool":            3.2,
	"technologies":    2.5,
	"technology":      3.0,
	"stack":           3.8,
	"infrastructure":  3.2,
	"security":        3.0,
	"authentication":  4.5,
	"authorization":   4.8,
	"encryption":      5.2,
	"oauth":           5.5,
	"saml":            6.0,
	"sso":             5.5,
	"compliance":      4.0,
	"gdpr":            5.5,
	"soc2":            6.0,
	"hipaa":           5.8,

	// ── Mobile ─────────────────────────────────────────────────────────
	"mobile":     3.5,
	"ios":        4.5,
	"android":    4.5,
	"flutter":    5.8,
	"native":     4.0,
	"responsive": 4.2,
	"app":        3.2,

	// ── Networking / Protocols ──────────────────────────────────────────
	"http":       4.5,
	"https":      5.0,
	"tcp":        5.5,
	"udp":        6.0,
	"websocket":  5.8,
	"grpc":       5.8,
	"protobuf":   6.2,
	"mqtt":       6.5,
	"dns":        5.5,
	"cdn":        5.5,
	"load":       3.8,
	"balancer":   5.2,
	"proxy":      5.2,
	"cache":      4.5,
	"caching":    4.8,
	"latency":    4.8,
	"throughput": 5.0,
	"bandwidth":  5.2,

	// ── Professional / soft skills ─────────────────────────────────────
	"leadership":     2.8,
	"management":     2.2,
	"communication":  2.5,
	"collaboration":  2.5,
	"mentoring":      3.8,
	"mentor":         4.2,
	"stakeholder":    3.2,
	"stakeholders":   3.2,
	"cross":          3.0,
	"problem":        2.8,
	"solving":        3.0,
	"critical":       3.5,
	"thinking":       3.8,
	"analytical":     3.5,
	"strategic":      3.2,
	"innovative":     3.5,
	"driven":         3.0,
	"motivated":      3.5,
	"passionate":     3.5,
	"detail":         3.2,
	"ownership":      3.5,
	"initiative":     3.5,
	"prioritize":     3.8,
	"prioritization": 4.2,
	"deadline":       3.8,
	"deadlines":      3.5,
	"fast":           3.5,
	"paced":          3.5,
	"dynamic":        3.2,

	// ── Business / Product ─────────────────────────────────────────────
	"product":     2.5,
	"business":    2.2,
	"revenue":     3.5,
	"growth":      3.0,
	"strategy":    3.0,
	"roadmap":     4.0,
	"metrics":     3.8,
	"kpi":         4.5,
	"roi":         4.8,
	"customer":    2.8,
	"customers":   3.0,
	"user":        3.0,
	"users":       3.2,
	"engagement":  3.5,
	"retention":   4.0,
	"conversion":  4.2,
	"acquisition": 4.0,
	"startup":     4.5,
	"enterprise":  3.5,
	"saas":        4.8,
	"b2b":         4.8,
	"b2c":         5.0,
	"ecommerce":   5.2,
	"fintech":     5.5,
	"healthtech":  6.0,
	"edtech":      6.2,

	// ── Education / Qualifications ─────────────────────────────────────
	"bachelor":      3.0,
	"bachelors":     3.2,
	"master":        3.5,
	"masters":       3.5,
	"phd":           4.5,
	"degree":        2.8,
	"university":    3.5,
	"certification": 3.8,
	"certified":     4.0,
	"years":         1.8,
	"year":          2.2,
	"senior":        2.5,
	"junior":        3.8,
	"mid":           4.0,
	"level":         3.0,
	"lead":          2.8,
	"principal":     4.5,
	"staff":         3.5,
	"director":      3.5,
	"vp":            4.8,
	"cto":           5.5,
	"ceo":           5.5,
	"manager":       2.5,
	"intern":        4.5,
	"internship":    4.5,
	"entry":         3.8,
	"experienced":   2.8,

	// ── Common resume action verbs ─────────────────────────────────────
	"built":        3.2,
	"designed":     3.0,
	"developed":    2.5,
	"implemented":  2.8,
	"managed":      2.8,
	"led":          2.8,
	"created":      3.0,
	"improved":     3.0,
	"increased":    3.2,
	"reduced":      3.2,
	"optimized":    3.5,
	"delivered":    3.0,
	"launched":     3.5,
	"maintained":   3.0,
	"collaborated": 3.0,
	"mentored":     4.0,
	"architected":  5.0,
	"migrated":     4.5,
	"scaled":       4.2,
	"deployed":     3.8,
	"integrated":   3.2,
	"resolved":     3.5,
	"analyzed":     3.2,
	"researched":   3.5,
	"presented":    3.5,
	"published":    4.2,
	"contributed":  3.5,
	"spearheaded":  4.5,
	"streamlined":  4.2,
	"established":  3.2,
	"transformed":  4.0,
	"coordinated":  3.2,
	"facilitated":  3.5,
	"negotiated":   4.0,

	// ── Miscellaneous tech terms ───────────────────────────────────────
	"blockchain":       5.8,
	"crypto":           6.0,
	"web3":             6.5,
	"iot":              5.5,
	"robotics":         6.0,
	"embedded":         5.5,
	"firmware":         6.0,
	"rtos":             7.0,
	"fpga":             7.0,
	"verilog":          7.5,
	"networking":       4.2,
	"virtualization":   5.2,
	"containerization": 5.5,
	"observability":    5.5,
	"sre":              5.5,
	"devsecops":        6.0,
	"chaos":            6.5,
	"mesh":             5.8,
	"gateway":          5.0,
	"queue":            4.8,
	"messaging":        4.8,
	"rabbitmq":         5.8,
	"sqs":              5.8,
	"sns":              6.0,
	"pubsub":           6.0,
	"event":            3.8,
	"streaming":        4.8,
	"batch":            4.5,
	"realtime":         5.2,
	"async":            5.2,
	"synchronous":      5.5,
	"asynchronous":     5.5,
	"multithreading":   5.8,
	"regex":            5.8,
	"json":             4.2,
	"xml":              4.8,
	"yaml":             5.2,
	"csv":              5.2,
	"pdf":              5.0,

	// ── Mechanical / Electrical / Industrial / Aerospace ──────────────
	"solidworks":      7.5,
	"catia":           7.8,
	"autocad":         6.2,
	"ansys":           6.8,
	"fea":             7.2,
	"cfd":             7.2,
	"gdt":             7.4,
	"tolerance":       5.8,
	"machining":       6.0,
	"cnc":             6.2,
	"fmea":            6.4,
	"sixsigma":        6.2,
	"plm":             6.3,
	"pneumatics":      6.3,
	"hydraulics":      5.8,
	"pcb":             6.9,
	"altium":          7.5,
	"kicad":           7.5,
	"microcontroller": 6.5,
	"plc":             6.6,
	"scada":           6.7,
	"inverter":        6.4,
	"simulink":        6.2,
	"oscilloscope":    6.4,
	"lean":            5.8,
	"kaizen":          6.6,
	"oee":             6.6,
	"supplychain":     5.5,
	"logistics":       4.8,
	"inventory":       4.8,
	"warehouse":       4.8,
	"procurement":     5.2,
	"forecasting":     5.0,
	"erp":             4.8,
	"sap":             5.2,
	"wms":             5.8,
	"aerodynamics":    7.4,
	"propulsion":      7.2,
	"avionics":        7.4,
	"aircraft":        6.0,
	"spacecraft":      7.6,
	"satellite":       6.4,
	"orbital":         7.0,
	"arinc":           7.8,
	"do178":           8.0,
	"do254":           8.0,
	"verification":    4.2,
	"validation":      4.0,
	"safety":          3.8,

	// ── Missing canonical terms from discipline profiles ──────────────
	"nx":              6.8,
	"thermodynamics":  6.2,
	"mechanics":       4.5,
	"materials":       3.8,
	"manufacturing":   3.5,
	"circuit":         5.2,
	"vhdl":            7.2,
	"power":           3.0,
	"signal":          4.0,
	"electronics":     4.5,
	"instrumentation": 5.8,
	"demand":          3.5,
	"planning":        2.8,
	"operations":      2.5,
	"transport":       4.8,
	"flight":          5.5,

	// ── Modern tech terms (2025-2026) ─────────────────────────────────
	"fastapi":      6.2,
	"playwright":   6.5,
	"prisma":       6.2,
	"vitest":       6.8,
	"bun":          6.5,
	"htmx":         6.8,
	"supabase":     6.2,
	"deno":         6.5,
	"astro":        6.5,
	"remix":        6.2,
	"solid":        5.8,
	"qwik":         7.2,
	"turbo":        6.0,
	"turborepo":    6.5,
	"drizzle":      7.0,
	"trpc":         6.8,
	"zod":          6.5,
	"tanstack":     6.8,
	"storybook":    5.8,
	"cypress":      5.8,
	"selenium":     5.2,
	"jest":         5.5,
	"mocha":        6.0,
	"pnpm":         6.2,
	"yarn":         5.5,
	"npm":          4.8,
	"openai":       6.0,
	"langchain":    6.5,
	"huggingface":  6.2,
	"mlflow":       6.5,
	"wandb":        6.8,
	"dbt":          6.2,
	"snowflake":    5.5,
	"bigquery":     5.8,
	"databricks":   5.8,
	"pulumi":       6.5,
	"nix":          6.8,
	"wasm":         6.5,
	"webassembly":  6.5,
	"temporal":     6.2,
	"nats":         6.5,
	"cockroachdb":  6.8,
	"planetscale":  6.8,
	"vercel":       6.0,
	"netlify":      6.0,
	"cloudflare":   5.5,

	// ── UX / Design ────────────────────────────────────────────────────
	"figma":         5.2,
	"sketch":        5.8,
	"adobe":         5.0,
	"photoshop":     5.8,
	"illustrator":   6.0,
	"ui":            3.8,
	"ux":            3.8,
	"wireframe":     5.5,
	"prototype":     4.5,
	"usability":     5.0,
	"accessibility": 4.5,
	"wcag":          6.2,
	"aria":          6.2,

	// ── French equivalents of common English terms ─────────────────────
	// These map French technical/professional vocabulary to the same IDF
	// values as their English counterparts, so they don't get the default
	// 7.0 ("rare/important") score. All accent-stripped.
	"developpement":   1.8, // développement = development
	"developpeur":     2.5, // développeur = developer
	"logiciel":        2.0, // = software
	"logiciels":       2.2, // = softwares
	"ingenieur":       2.2, // ingénieur = engineer
	"ingenierie":      2.5, // ingénierie = engineering
	"informatique":    2.5, // = computer science / IT
	"conception":      2.5, // = design
	"gestion":         2.2, // = management
	"securite":        3.0, // sécurité = security
	"donnees":         2.0, // données = data
	"reseau":          4.2, // réseau = network
	"reseaux":         4.2, // réseaux = networks
	"deploiement":     3.5, // déploiement = deployment
	"amelioration":    3.0, // amélioration = improvement
	"fonctionnalite":  3.5, // fonctionnalité = functionality
	"fonctionnalites": 3.5,
	"automatisation":  3.5, // = automation
	"optimisation":    3.8, // = optimization
	"maintenance":     3.0, // = maintenance
	"mise":            2.5, // (as in "mise en place", "mise en production")
	"place":           3.0,
	"oeuvre":          3.5, // (as in "mise en oeuvre")
	"utilisateur":     3.0, // = user
	"utilisateurs":    3.2,
	"client":          2.8, // = client/customer
	"clients":         3.0,
	"serveur":         4.0, // = server
	"base":            3.0, // (as in "base de données")
	"test":            3.2, // = test
	"tests":           3.2,
	"unitaire":        3.8, // = unit (test)
	"unitaires":       3.8,
	"automatise":      3.8, // = automated (tests automatisés)
	// "integration" already in English section (3.2)
	"equipe":     2.5, // équipe = team (low IDF, generic)
	"projet":     2.5, // = project
	"projets":    2.5,
	"technique":  2.5, // = technical
	"techniques": 2.5,
	"methode":    3.5, // méthode = method
	"methodes":   3.5,
	// "agile" already in English section (3.5)
	"performant":   3.8, // = performant/high-performance
	"performante":  3.8,
	"qualite":      3.2, // qualité = quality
	"outil":        3.2, // = tool
	"outils":       2.8, // = tools
	"formation":    3.5, // = education/training
	"diplome":      2.8, // diplôme = degree
	"licence":      3.5, // = bachelor's degree (FR)
	"maitrise":     3.5, // maîtrise = master's degree (FR/QC)
	"baccalaureat": 3.0, // baccalauréat = bachelor's (QC) / high school diploma (FR)
}

// phraseIDF maps multi-word phrases (space-separated, lowercased) to their IDF values.
// These are curated known phrases — not adjacent bigrams — to avoid false positives.
var phraseIDF = map[string]float64{
	// ── AI / ML ────────────────────────────────────────────────────────
	"machine learning":       4.0,
	"deep learning":          5.0,
	"natural language":       5.2,
	"natural language processing": 5.5,
	"computer vision":        5.5,
	"neural network":         5.2,
	"neural networks":        5.2,
	"large language":         5.8,
	"reinforcement learning": 6.0,
	"transfer learning":      6.2,
	// ── Data ───────────────────────────────────────────────────────────
	"data science":     4.2,
	"data engineering": 5.2,
	"data analysis":    4.0,
	"data analytics":   4.2,
	// ── Cloud / Infrastructure ──────────────────────────────────────────
	"cloud native":     5.2,
	"cloud computing":  3.8,
	"high availability": 5.0,
	"load balancing":   5.2,
	"message queue":    5.5,
	"service mesh":     5.8,
	"infrastructure as": 5.5,
	// ── Engineering practices ───────────────────────────────────────────
	"full stack":             4.5,
	"front end":              3.5,
	"back end":               3.5,
	"continuous integration": 4.8,
	"continuous delivery":    5.0,
	"test driven":            5.5,
	"domain driven":          5.8,
	"object oriented":        4.5,
	"distributed systems":    5.0,
	"real time":              4.5,
	"event driven":           5.2,
	"software engineering":   2.8,
	"software development":   2.5,
	"version control":        4.2,
	"code review":            4.0,
	// ── Business ───────────────────────────────────────────────────────
	"product management": 3.5,
	"project management": 3.2,
	// ── Mechanical / Electrical / Industrial ───────────────────────────
	"finite element":     6.5,
	"computational fluid": 6.8,
	"fluid dynamics":     6.5,
	"heat transfer":      6.5,
	"printed circuit":    6.5,
	"embedded systems":   5.8,
	"signal processing":  6.0,
	"power electronics":  6.5,
	"supply chain":       5.5,
	"lean manufacturing": 6.0,
	"quality assurance":  5.0,
	"failure mode":       6.5,
}

// phraseCategories maps each phrase key to its scoring bucket.
var phraseCategories = map[string]string{
	// AI / ML → practices
	"machine learning":            categoryPractices,
	"deep learning":               categoryPractices,
	"natural language":            categoryPractices,
	"natural language processing": categoryPractices,
	"computer vision":             categoryPractices,
	"neural network":              categoryPractices,
	"neural networks":             categoryPractices,
	"large language":              categoryPractices,
	"reinforcement learning":      categoryPractices,
	"transfer learning":           categoryPractices,
	// Data → practices
	"data science":     categoryPractices,
	"data engineering": categoryPractices,
	"data analysis":    categoryPractices,
	"data analytics":   categoryPractices,
	// Cloud / Infrastructure → cloud_devops_db
	"cloud native":      categoryCloudDevOps,
	"cloud computing":   categoryCloudDevOps,
	"high availability": categoryCloudDevOps,
	"load balancing":    categoryCloudDevOps,
	"message queue":     categoryCloudDevOps,
	"service mesh":      categoryCloudDevOps,
	"infrastructure as": categoryCloudDevOps,
	// Engineering practices → practices
	"full stack":             categoryPractices,
	"front end":              categoryPractices,
	"back end":               categoryPractices,
	"continuous integration": categoryPractices,
	"continuous delivery":    categoryPractices,
	"test driven":            categoryPractices,
	"domain driven":          categoryPractices,
	"object oriented":        categoryPractices,
	"distributed systems":    categoryPractices,
	"real time":              categoryPractices,
	"event driven":           categoryPractices,
	"software engineering":   categoryPractices,
	"software development":   categoryPractices,
	"version control":        categoryPractices,
	"code review":            categoryPractices,
	// Business → soft_skills
	"product management": categorySoftSkills,
	"project management": categorySoftSkills,
	// Mechanical / Electrical / Industrial
	"finite element":      "simulation_analysis",
	"computational fluid": "simulation_analysis",
	"fluid dynamics":      "simulation_analysis",
	"heat transfer":       "simulation_analysis",
	"printed circuit":     "electronics",
	"embedded systems":    "electronics",
	"signal processing":   "electronics",
	"power electronics":   "electronics",
	"supply chain":        "operations",
	"lean manufacturing":  "operations",
	"quality assurance":   "operations",
	"failure mode":        "simulation_analysis",
}

// lookupIDF returns the IDF value for a given term from the static table.
// It also checks canonical synonym forms. Unknown terms receive defaultCorpusIDF.
func lookupIDF(term string) float64 {
	if v, ok := corpusIDF[term]; ok {
		return v
	}
	if v, ok := phraseIDF[term]; ok {
		return v
	}
	// Try the canonical form (e.g. "go" → "golang")
	if canon := canonicalize(term); canon != term {
		if v, ok := corpusIDF[canon]; ok {
			return v
		}
	}
	if isLikelyNoisyToken(term) {
		return noisyUnknownIDF
	}
	if len(term) <= 3 {
		return shortUnknownIDF
	}
	return defaultCorpusIDF
}

func isLikelyNoisyToken(term string) bool {
	if len(term) > 24 {
		return true
	}

	letters := 0
	digits := 0
	vowels := 0
	repeatRun := 1
	var prev rune

	for i, r := range term {
		if unicode.IsLetter(r) {
			letters++
			if strings.ContainsRune("aeiouy", r) {
				vowels++
			}
		}
		if unicode.IsDigit(r) {
			digits++
		}
		if i > 0 {
			if r == prev {
				repeatRun++
				if repeatRun >= 4 {
					return true
				}
			} else {
				repeatRun = 1
			}
		}
		prev = r
	}

	total := letters + digits
	if letters > 0 && digits > 0 && digits*2 >= total {
		return true
	}
	if letters >= 10 && vowels == 0 {
		return true
	}

	return false
}
