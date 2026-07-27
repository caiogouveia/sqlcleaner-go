package main

// Dumps de exemplo usados nos testes, espelhando fielmente as constantes
// DUMP_SAMPLE/DUMP_EVENTS/DUMP_CASE_SENSITIVE/DUMP_LITERAL_DOT/DUMP_SET_SAMPLE
// de test_scripts.py no repositório Python original.

const dumpSample = "--\n" +
	"-- PostgreSQL database dump\n" +
	"--\n" +
	"\n" +
	"SET statement_timeout = 0;\n" +
	"\n" +
	"CREATE TABLE public.users (\n" +
	"    id integer NOT NULL,\n" +
	"    name text\n" +
	");\n" +
	"\n" +
	"CREATE SEQUENCE public.users_id_seq\n" +
	"    START WITH 1\n" +
	"    INCREMENT BY 1;\n" +
	"\n" +
	"ALTER TABLE public.users ALTER COLUMN id SET DEFAULT nextval('public.users_id_seq'::regclass);\n" +
	"\n" +
	"CREATE INDEX idx_users_name ON public.users USING btree (name);\n" +
	"\n" +
	"CREATE TABLE public.logs (\n" +
	"    id integer NOT NULL,\n" +
	"    message text\n" +
	");\n" +
	"\n" +
	"CREATE SEQUENCE public.logs_id_seq\n" +
	"    START WITH 1\n" +
	"    INCREMENT BY 1;\n" +
	"\n" +
	"ALTER TABLE public.logs ALTER COLUMN id SET DEFAULT nextval('public.logs_id_seq'::regclass);\n" +
	"\n" +
	"CREATE INDEX idx_logs_id ON public.logs USING btree (id);\n" +
	"\n" +
	"COPY public.users (id, name) FROM stdin;\n" +
	"1\tAlice\n" +
	"2\tBob\n" +
	"\\.\n" +
	"\n" +
	"COPY public.logs (id, message) FROM stdin;\n" +
	"1\tstartup\n" +
	"2\tshutdown\n" +
	"\\.\n" +
	"\n" +
	"ALTER TABLE ONLY public.users\n" +
	"    ADD CONSTRAINT users_pkey PRIMARY KEY (id);\n" +
	"\n" +
	"ALTER TABLE ONLY public.logs\n" +
	"    ADD CONSTRAINT logs_pkey PRIMARY KEY (id);\n"

const dumpEvents = "CREATE TABLE public.events (\n" +
	"    id integer NOT NULL,\n" +
	"    payload text\n" +
	");\n" +
	"\n" +
	"COPY public.events (id, payload) FROM stdin;\n" +
	"1\te1\n" +
	"2\te2\n" +
	"3\te3\n" +
	"4\te4\n" +
	"5\te5\n" +
	"\\.\n" +
	"\n" +
	"CREATE TABLE public.empty_table (\n" +
	"    id integer NOT NULL\n" +
	");\n" +
	"\n" +
	"COPY public.empty_table (id) FROM stdin;\n" +
	"\\.\n"

const dumpCaseSensitive = "CREATE TABLE public.\"ticketsDW\" (\n" +
	"    id integer NOT NULL\n" +
	");\n" +
	"\n" +
	"CREATE SEQUENCE public.\"ticketsDW_id_seq\"\n" +
	"    START WITH 1\n" +
	"    INCREMENT BY 1;\n" +
	"\n" +
	"ALTER TABLE ONLY public.\"ticketsDW\"\n" +
	"    ALTER COLUMN id SET DEFAULT nextval('public.\"ticketsDW_id_seq\"'::regclass);\n" +
	"\n" +
	"CREATE INDEX idx_ticketsdw_id ON public.\"ticketsDW\" USING btree (id);\n" +
	"\n" +
	"COPY public.\"ticketsDW\" (id) FROM stdin;\n" +
	"1\n" +
	"2\n" +
	"\\.\n" +
	"\n" +
	"CREATE TABLE \"DW\".\"ticketsDW\" (\n" +
	"    id integer NOT NULL\n" +
	");\n" +
	"\n" +
	"COPY \"DW\".\"ticketsDW\" (id) FROM stdin;\n" +
	"10\n" +
	"\\.\n"

const dumpLiteralDot = "CREATE TABLE public.\"DW.ticketsDW\" (\n" +
	"    id integer NOT NULL\n" +
	");\n" +
	"\n" +
	"COPY public.\"DW.ticketsDW\" (id) FROM stdin;\n" +
	"1\n" +
	"2\n" +
	"\\.\n" +
	"\n" +
	"CREATE TABLE public.outra (\n" +
	"    id integer NOT NULL\n" +
	");\n" +
	"\n" +
	"COPY public.outra (id) FROM stdin;\n" +
	"9\n" +
	"\\.\n"

const dumpSetSample = "SET transaction_timeout = 0;\n" +
	"SET statement_timeout = 0;\n" +
	"set TRANSACTION_TIMEOUT = 5000;\n" +
	"SET \"transaction_timeout\" TO 3000;\n" +
	"SET transaction_timeout_like = 1;\n" +
	"SET search_path = public;\n" +
	"\n" +
	"CREATE TABLE public.users (\n" +
	"    id integer NOT NULL\n" +
	");\n" +
	"\n" +
	"COPY public.users (id) FROM stdin;\n" +
	"1\n" +
	"\\.\n"
