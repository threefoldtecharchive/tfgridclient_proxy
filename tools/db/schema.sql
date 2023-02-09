--
-- PostgreSQL database dump
--

-- Dumped from database version 12.9 (Debian 12.9-1.pgdg110+1)
-- Dumped by pg_dump version 12.9 (Debian 12.9-1.pgdg110+1)

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: substrate_threefold_status; Type: SCHEMA; Schema: -; Owner: postgres
--

CREATE SCHEMA substrate_threefold_status;


ALTER SCHEMA substrate_threefold_status OWNER TO postgres;

--
-- Name: node_resources(integer); Type: FUNCTION; Schema: public; Owner: postgres
--

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: account; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.account (
    id character varying NOT NULL,
    wallet text NOT NULL,
    balance numeric NOT NULL
);


ALTER TABLE public.account OWNER TO postgres;

--
-- Name: burn_transaction; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.burn_transaction (
    id character varying NOT NULL,
    block integer NOT NULL,
    amount numeric NOT NULL,
    target text NOT NULL
);


ALTER TABLE public.burn_transaction OWNER TO postgres;

--
-- Name: city; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.city (
    id character varying NOT NULL,
    city_id integer NOT NULL,
    country_id integer NOT NULL,
    name text NOT NULL
);


ALTER TABLE public.city OWNER TO postgres;

--
-- Name: contract_bill_report; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.contract_bill_report (
    id character varying NOT NULL,
    contract_id numeric NOT NULL,
    discount_received character varying(7) NOT NULL,
    amount_billed numeric NOT NULL,
    "timestamp" numeric NOT NULL
);


ALTER TABLE public.contract_bill_report OWNER TO postgres;

--
-- Name: contract_resources; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.contract_resources (
    id character varying NOT NULL,
    hru numeric NOT NULL,
    sru numeric NOT NULL,
    cru numeric NOT NULL,
    mru numeric NOT NULL,
    contract_id character varying NOT NULL
);


ALTER TABLE public.contract_resources OWNER TO postgres;

--
-- Name: country; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.country (
    id character varying NOT NULL,
    country_id integer NOT NULL,
    code text NOT NULL,
    name text NOT NULL,
    region text NOT NULL,
    subregion text NOT NULL,
    lat text,
    long text
);


ALTER TABLE public.country OWNER TO postgres;

--
-- Name: entity; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.entity (
    id character varying NOT NULL,
    grid_version integer NOT NULL,
    entity_id integer NOT NULL,
    name text NOT NULL,
    country text,
    city text,
    account_id text NOT NULL
);


ALTER TABLE public.entity OWNER TO postgres;

--
-- Name: entity_proof; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.entity_proof (
    id character varying NOT NULL,
    entity_id integer NOT NULL,
    signature text NOT NULL,
    twin_rel_id character varying NOT NULL
);


ALTER TABLE public.entity_proof OWNER TO postgres;

--
-- Name: farm; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.farm (
    id character varying NOT NULL,
    grid_version integer NOT NULL,
    farm_id integer NOT NULL,
    name text NOT NULL,
    twin_id integer NOT NULL,
    pricing_policy_id integer NOT NULL,
    stellar_address text,
    dedicated_farm boolean,
    certification character varying(12)
);


ALTER TABLE public.farm OWNER TO postgres;

--
-- Name: farming_policy; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.farming_policy (
    id character varying NOT NULL,
    grid_version integer NOT NULL,
    farming_policy_id integer NOT NULL,
    name text,
    cu integer,
    su integer,
    nu integer,
    ipv4 integer,
    minimal_uptime integer,
    policy_created integer,
    policy_end integer,
    immutable boolean,
    "default" boolean,
    node_certification character varying(9),
    farm_certification character varying(12)
);


ALTER TABLE public.farming_policy OWNER TO postgres;

--
-- Name: historical_balance; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.historical_balance (
    id character varying NOT NULL,
    balance numeric NOT NULL,
    "timestamp" numeric NOT NULL,
    account_id character varying NOT NULL
);


ALTER TABLE public.historical_balance OWNER TO postgres;

--
-- Name: interfaces; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.interfaces (
    id character varying NOT NULL,
    name text NOT NULL,
    mac text NOT NULL,
    ips text NOT NULL,
    node_id character varying NOT NULL
);


ALTER TABLE public.interfaces OWNER TO postgres;

--
-- Name: location; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.location (
    id character varying NOT NULL,
    longitude text NOT NULL,
    latitude text NOT NULL
);


ALTER TABLE public.location OWNER TO postgres;

--
-- Name: migrations; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.migrations (
    id integer NOT NULL,
    "timestamp" bigint NOT NULL,
    name character varying NOT NULL
);


ALTER TABLE public.migrations OWNER TO postgres;

--
-- Name: migrations_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.migrations_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.migrations_id_seq OWNER TO postgres;

--
-- Name: migrations_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.migrations_id_seq OWNED BY public.migrations.id;


--
-- Name: mint_transaction; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.mint_transaction (
    id character varying NOT NULL,
    amount numeric NOT NULL,
    target text NOT NULL,
    block integer NOT NULL
);


ALTER TABLE public.mint_transaction OWNER TO postgres;

--
-- Name: name_contract; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.name_contract (
    id character varying NOT NULL,
    grid_version integer NOT NULL,
    contract_id numeric NOT NULL,
    twin_id integer NOT NULL,
    name text NOT NULL,
    created_at numeric NOT NULL,
    state character varying(11) NOT NULL
);


ALTER TABLE public.name_contract OWNER TO postgres;

--
-- Name: node; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.node (
    id character varying NOT NULL,
    grid_version integer NOT NULL,
    node_id integer NOT NULL,
    farm_id integer NOT NULL,
    twin_id integer NOT NULL,
    country text,
    city text,
    uptime numeric,
    created integer NOT NULL,
    farming_policy_id integer NOT NULL,
    secure boolean,
    virtualized boolean,
    serial_number text,
    created_at numeric NOT NULL,
    updated_at numeric NOT NULL,
    location_id character varying NOT NULL,
    certification character varying(9),
    connection_price integer
);


ALTER TABLE public.node OWNER TO postgres;

--
-- Name: node_contract; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.node_contract (
    id character varying NOT NULL,
    grid_version integer NOT NULL,
    contract_id numeric NOT NULL,
    twin_id integer NOT NULL,
    node_id integer NOT NULL,
    deployment_data text NOT NULL,
    deployment_hash text NOT NULL,
    number_of_public_i_ps integer NOT NULL,
    created_at numeric NOT NULL,
    resources_used_id character varying,
    state character varying(11) NOT NULL
);


ALTER TABLE public.node_contract OWNER TO postgres;

--
-- Name: node_resources_free; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.node_resources_free (
    id character varying NOT NULL,
    hru numeric NOT NULL,
    sru numeric NOT NULL,
    cru numeric NOT NULL,
    mru numeric NOT NULL,
    node_id character varying NOT NULL
);


ALTER TABLE public.node_resources_free OWNER TO postgres;

--
-- Name: node_resources_total; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.node_resources_total (
    id character varying NOT NULL,
    hru numeric NOT NULL,
    sru numeric NOT NULL,
    cru numeric NOT NULL,
    mru numeric NOT NULL,
    node_id character varying NOT NULL
);


ALTER TABLE public.node_resources_total OWNER TO postgres;

--
-- Name: node_resources_used; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.node_resources_used (
    id character varying NOT NULL,
    hru numeric NOT NULL,
    sru numeric NOT NULL,
    cru numeric NOT NULL,
    mru numeric NOT NULL,
    node_id character varying NOT NULL
);


ALTER TABLE public.node_resources_used OWNER TO postgres;

--
-- Name: nodes_resources_view; Type: VIEW; Schema: public; Owner: postgres
--

CREATE VIEW public.nodes_resources_view AS
 SELECT node.node_id,
    COALESCE(sum(contract_resources.cru), (0)::numeric) AS used_cru,
    (COALESCE(sum(contract_resources.mru), (0)::numeric) + (GREATEST(((node_resources_total.mru / (10)::numeric))::bigint, '2147483648'::bigint))::numeric) AS used_mru,
    COALESCE(sum(contract_resources.hru), (0)::numeric) AS used_hru,
    (COALESCE(sum(contract_resources.sru), (0)::numeric) + ('107374182400'::bigint)::numeric) AS used_sru,
    ((node_resources_total.mru - COALESCE(sum(contract_resources.mru), (0)::numeric)) - (GREATEST(((node_resources_total.mru / (10)::numeric))::bigint, '2147483648'::bigint))::numeric) AS free_mru,
    (node_resources_total.hru - COALESCE(sum(contract_resources.hru), (0)::numeric)) AS free_hru,
    ((node_resources_total.sru - COALESCE(sum(contract_resources.sru), (0)::numeric)) - ('107374182400'::bigint)::numeric) AS free_sru,
    COALESCE(node_resources_total.cru, (0)::numeric) AS total_cru,
    COALESCE(node_resources_total.mru, (0)::numeric) AS total_mru,
    COALESCE(node_resources_total.hru, (0)::numeric) AS total_hru,
    COALESCE(node_resources_total.sru, (0)::numeric) AS total_sru
   FROM (((public.contract_resources
     JOIN public.node_contract node_contract ON ((((node_contract.resources_used_id)::text = (contract_resources.id)::text) AND ((node_contract.state)::text = 'Created'::text))))
     RIGHT JOIN public.node node ON ((node.node_id = node_contract.node_id)))
     JOIN public.node_resources_total node_resources_total ON (((node_resources_total.node_id)::text = (node.id)::text)))
  GROUP BY node.node_id, node_resources_total.mru, node_resources_total.sru, node_resources_total.hru, node_resources_total.cru;


ALTER TABLE public.nodes_resources_view OWNER TO postgres;

--
-- Name: nru_consumption; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.nru_consumption (
    id character varying NOT NULL,
    contract_id numeric NOT NULL,
    "timestamp" numeric NOT NULL,
    "window" numeric,
    nru numeric
);


ALTER TABLE public.nru_consumption OWNER TO postgres;

--
-- Name: pricing_policy; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.pricing_policy (
    id character varying NOT NULL,
    grid_version integer NOT NULL,
    pricing_policy_id integer NOT NULL,
    name text NOT NULL,
    su jsonb NOT NULL,
    cu jsonb NOT NULL,
    nu jsonb NOT NULL,
    ipu jsonb NOT NULL,
    foundation_account text NOT NULL,
    certified_sales_account text NOT NULL,
    dedicated_node_discount integer NOT NULL
);


ALTER TABLE public.pricing_policy OWNER TO postgres;

--
-- Name: public_config; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.public_config (
    id character varying NOT NULL,
    ipv4 text,
    ipv6 text,
    gw4 text,
    gw6 text,
    domain text,
    node_id character varying NOT NULL
);


ALTER TABLE public.public_config OWNER TO postgres;

--
-- Name: public_ip; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.public_ip (
    id character varying NOT NULL,
    gateway text NOT NULL,
    ip text NOT NULL,
    contract_id numeric NOT NULL,
    farm_id character varying NOT NULL
);


ALTER TABLE public.public_ip OWNER TO postgres;

--
-- Name: refund_transaction; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.refund_transaction (
    id character varying NOT NULL,
    block integer NOT NULL,
    amount numeric NOT NULL,
    target text NOT NULL,
    tx_hash text NOT NULL
);


ALTER TABLE public.refund_transaction OWNER TO postgres;

--
-- Name: rent_contract; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.rent_contract (
    id character varying NOT NULL,
    grid_version integer NOT NULL,
    contract_id numeric NOT NULL,
    twin_id integer NOT NULL,
    node_id integer NOT NULL,
    created_at numeric NOT NULL,
    state character varying(11) NOT NULL
);


ALTER TABLE public.rent_contract OWNER TO postgres;

--
-- Name: transfer; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.transfer (
    id character varying NOT NULL,
    "from" text NOT NULL,
    "to" text NOT NULL,
    amount numeric NOT NULL,
    "timestamp" numeric NOT NULL
);


ALTER TABLE public.transfer OWNER TO postgres;

--
-- Name: twin; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.twin (
    id character varying NOT NULL,
    grid_version integer NOT NULL,
    twin_id integer NOT NULL,
    account_id text NOT NULL,
    relay text,
    public_key text
);


ALTER TABLE public.twin OWNER TO postgres;

--
-- Name: typeorm_metadata; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.typeorm_metadata (
    type character varying NOT NULL,
    database character varying,
    schema character varying,
    "table" character varying,
    name character varying,
    value text
);


ALTER TABLE public.typeorm_metadata OWNER TO postgres;

--
-- Name: uptime_event; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.uptime_event (
    id character varying NOT NULL,
    node_id integer NOT NULL,
    uptime numeric NOT NULL,
    "timestamp" numeric NOT NULL
);


ALTER TABLE public.uptime_event OWNER TO postgres;

--
-- Name: status; Type: TABLE; Schema: substrate_threefold_status; Owner: postgres
--

CREATE TABLE substrate_threefold_status.status (
    id integer NOT NULL,
    height integer NOT NULL
);


ALTER TABLE substrate_threefold_status.status OWNER TO postgres;

--
-- Name: migrations id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.migrations ALTER COLUMN id SET DEFAULT nextval('public.migrations_id_seq'::regclass);


--
-- Name: node_resources_used PK_05bf9bc81d419c0f34c8bf08d5f; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.node_resources_used
    ADD CONSTRAINT "PK_05bf9bc81d419c0f34c8bf08d5f" PRIMARY KEY (id);


--
-- Name: node_resources_free PK_0a15fb3f274365eef34123c2dea; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.node_resources_free
    ADD CONSTRAINT "PK_0a15fb3f274365eef34123c2dea" PRIMARY KEY (id);


--
-- Name: twin PK_18457170fa91d0a787d9f635d7c; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.twin
    ADD CONSTRAINT "PK_18457170fa91d0a787d9f635d7c" PRIMARY KEY (id);


--
-- Name: mint_transaction PK_19f4328320501dfd14e2bae0855; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.mint_transaction
    ADD CONSTRAINT "PK_19f4328320501dfd14e2bae0855" PRIMARY KEY (id);


--
-- Name: burn_transaction PK_20ec76c5c56dd6b47dec5f0aaa8; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.burn_transaction
    ADD CONSTRAINT "PK_20ec76c5c56dd6b47dec5f0aaa8" PRIMARY KEY (id);


--
-- Name: farm PK_3bf246b27a3b6678dfc0b7a3f64; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.farm
    ADD CONSTRAINT "PK_3bf246b27a3b6678dfc0b7a3f64" PRIMARY KEY (id);


--
-- Name: rent_contract PK_3c99766b627604d5950d704e33a; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.rent_contract
    ADD CONSTRAINT "PK_3c99766b627604d5950d704e33a" PRIMARY KEY (id);


--
-- Name: entity PK_50a7741b415bc585fcf9c984332; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.entity
    ADD CONSTRAINT "PK_50a7741b415bc585fcf9c984332" PRIMARY KEY (id);


--
-- Name: account PK_54115ee388cdb6d86bb4bf5b2ea; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.account
    ADD CONSTRAINT "PK_54115ee388cdb6d86bb4bf5b2ea" PRIMARY KEY (id);


--
-- Name: contract_resources PK_557de19994fcca90916e8c6582f; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.contract_resources
    ADD CONSTRAINT "PK_557de19994fcca90916e8c6582f" PRIMARY KEY (id);


--
-- Name: contract_bill_report PK_5b21fd81e47bddc5f1fdbc8d7ee; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.contract_bill_report
    ADD CONSTRAINT "PK_5b21fd81e47bddc5f1fdbc8d7ee" PRIMARY KEY (id);


--
-- Name: farming_policy PK_5d2ec9534104f44e4d989c4e82f; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.farming_policy
    ADD CONSTRAINT "PK_5d2ec9534104f44e4d989c4e82f" PRIMARY KEY (id);


--
-- Name: historical_balance PK_74ac29ad0bdffb6d1281a1e17e8; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.historical_balance
    ADD CONSTRAINT "PK_74ac29ad0bdffb6d1281a1e17e8" PRIMARY KEY (id);


--
-- Name: refund_transaction PK_74ffc5427c595968dd777f71bf4; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.refund_transaction
    ADD CONSTRAINT "PK_74ffc5427c595968dd777f71bf4" PRIMARY KEY (id);


--
-- Name: pricing_policy PK_78105eb11bd75fd76a23bbc9bb1; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.pricing_policy
    ADD CONSTRAINT "PK_78105eb11bd75fd76a23bbc9bb1" PRIMARY KEY (id);


--
-- Name: public_config PK_7839f7dd8f45e37933fb3e35cbb; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.public_config
    ADD CONSTRAINT "PK_7839f7dd8f45e37933fb3e35cbb" PRIMARY KEY (id);


--
-- Name: name_contract PK_7b4cd056bbb83602d211996360f; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.name_contract
    ADD CONSTRAINT "PK_7b4cd056bbb83602d211996360f" PRIMARY KEY (id);


--
-- Name: interfaces PK_811ec6e568e3c1a89ac5e744731; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.interfaces
    ADD CONSTRAINT "PK_811ec6e568e3c1a89ac5e744731" PRIMARY KEY (id);


--
-- Name: location PK_876d7bdba03c72251ec4c2dc827; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.location
    ADD CONSTRAINT "PK_876d7bdba03c72251ec4c2dc827" PRIMARY KEY (id);


--
-- Name: migrations PK_8c82d7f526340ab734260ea46be; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.migrations
    ADD CONSTRAINT "PK_8c82d7f526340ab734260ea46be" PRIMARY KEY (id);


--
-- Name: node PK_8c8caf5f29d25264abe9eaf94dd; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.node
    ADD CONSTRAINT "PK_8c8caf5f29d25264abe9eaf94dd" PRIMARY KEY (id);


--
-- Name: uptime_event PK_90783463b0d0b660367ebd7f5ff; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.uptime_event
    ADD CONSTRAINT "PK_90783463b0d0b660367ebd7f5ff" PRIMARY KEY (id);


--
-- Name: node_resources_total PK_964127f256a8ffeba2aa31c098d; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.node_resources_total
    ADD CONSTRAINT "PK_964127f256a8ffeba2aa31c098d" PRIMARY KEY (id);


--
-- Name: node_contract PK_a5f90b17f504ffcd79d1f66574a; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.node_contract
    ADD CONSTRAINT "PK_a5f90b17f504ffcd79d1f66574a" PRIMARY KEY (id);


--
-- Name: city PK_b222f51ce26f7e5ca86944a6739; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.city
    ADD CONSTRAINT "PK_b222f51ce26f7e5ca86944a6739" PRIMARY KEY (id);


--
-- Name: entity_proof PK_b55dee5f461106682013d0beef8; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.entity_proof
    ADD CONSTRAINT "PK_b55dee5f461106682013d0beef8" PRIMARY KEY (id);


--
-- Name: country PK_bf6e37c231c4f4ea56dcd887269; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.country
    ADD CONSTRAINT "PK_bf6e37c231c4f4ea56dcd887269" PRIMARY KEY (id);


--
-- Name: nru_consumption PK_ca7956fb8fcdb7198737387d9a8; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.nru_consumption
    ADD CONSTRAINT "PK_ca7956fb8fcdb7198737387d9a8" PRIMARY KEY (id);


--
-- Name: public_ip PK_f170b0b519632730f41d2ef78f4; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.public_ip
    ADD CONSTRAINT "PK_f170b0b519632730f41d2ef78f4" PRIMARY KEY (id);


--
-- Name: transfer PK_fd9ddbdd49a17afcbe014401295; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.transfer
    ADD CONSTRAINT "PK_fd9ddbdd49a17afcbe014401295" PRIMARY KEY (id);


--
-- Name: node_resources_used REL_75870a8ed1c14efd1dd4ef4792; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.node_resources_used
    ADD CONSTRAINT "REL_75870a8ed1c14efd1dd4ef4792" UNIQUE (node_id);


--
-- Name: node_resources_free REL_923c4dff43306d0a0f5a98a1ab; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.node_resources_free
    ADD CONSTRAINT "REL_923c4dff43306d0a0f5a98a1ab" UNIQUE (node_id);


--
-- Name: public_config REL_d394b8b9afbb1b1a2346f9743c; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.public_config
    ADD CONSTRAINT "REL_d394b8b9afbb1b1a2346f9743c" UNIQUE (node_id);


--
-- Name: node_resources_total REL_fd430c3a2645c8f409f859c2aa; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.node_resources_total
    ADD CONSTRAINT "REL_fd430c3a2645c8f409f859c2aa" UNIQUE (node_id);


--
-- Name: status status_pkey; Type: CONSTRAINT; Schema: substrate_threefold_status; Owner: postgres
--

ALTER TABLE ONLY substrate_threefold_status.status
    ADD CONSTRAINT status_pkey PRIMARY KEY (id);


--
-- Name: IDX_23937641f28c607f061dab4694; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "IDX_23937641f28c607f061dab4694" ON public.interfaces USING btree (node_id);


--
-- Name: IDX_383ff006e4b59db91d32cb891e; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "IDX_383ff006e4b59db91d32cb891e" ON public.historical_balance USING btree (account_id);


--
-- Name: IDX_3d9cbf30c68b79a801e1d5c9b4; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "IDX_3d9cbf30c68b79a801e1d5c9b4" ON public.entity_proof USING btree (twin_rel_id);


--
-- Name: IDX_5cc2d1af1d8132b614abd340b0; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "IDX_5cc2d1af1d8132b614abd340b0" ON public.public_ip USING btree (farm_id);


--
-- Name: IDX_621238dffde9099b2233650235; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "IDX_621238dffde9099b2233650235" ON public.contract_resources USING btree (contract_id);


--
-- Name: IDX_75870a8ed1c14efd1dd4ef4792; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX "IDX_75870a8ed1c14efd1dd4ef4792" ON public.node_resources_used USING btree (node_id);


--
-- Name: IDX_923c4dff43306d0a0f5a98a1ab; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX "IDX_923c4dff43306d0a0f5a98a1ab" ON public.node_resources_free USING btree (node_id);


--
-- Name: IDX_d224b7b862841f24dd85b55605; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "IDX_d224b7b862841f24dd85b55605" ON public.node USING btree (location_id);


--
-- Name: IDX_d394b8b9afbb1b1a2346f9743c; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX "IDX_d394b8b9afbb1b1a2346f9743c" ON public.public_config USING btree (node_id);


--
-- Name: IDX_f294cfb50bb7c7b976d86c08fd; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "IDX_f294cfb50bb7c7b976d86c08fd" ON public.node_contract USING btree (resources_used_id);


--
-- Name: IDX_fd430c3a2645c8f409f859c2aa; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX "IDX_fd430c3a2645c8f409f859c2aa" ON public.node_resources_total USING btree (node_id);


--
-- Name: interfaces FK_23937641f28c607f061dab4694b; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.interfaces
    ADD CONSTRAINT "FK_23937641f28c607f061dab4694b" FOREIGN KEY (node_id) REFERENCES public.node(id);


--
-- Name: historical_balance FK_383ff006e4b59db91d32cb891e9; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.historical_balance
    ADD CONSTRAINT "FK_383ff006e4b59db91d32cb891e9" FOREIGN KEY (account_id) REFERENCES public.account(id);


--
-- Name: entity_proof FK_3d9cbf30c68b79a801e1d5c9b41; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.entity_proof
    ADD CONSTRAINT "FK_3d9cbf30c68b79a801e1d5c9b41" FOREIGN KEY (twin_rel_id) REFERENCES public.twin(id);


--
-- Name: public_ip FK_5cc2d1af1d8132b614abd340b06; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.public_ip
    ADD CONSTRAINT "FK_5cc2d1af1d8132b614abd340b06" FOREIGN KEY (farm_id) REFERENCES public.farm(id);


--
-- Name: contract_resources FK_621238dffde9099b2233650235d; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.contract_resources
    ADD CONSTRAINT "FK_621238dffde9099b2233650235d" FOREIGN KEY (contract_id) REFERENCES public.node_contract(id);


--
-- Name: node_resources_used FK_75870a8ed1c14efd1dd4ef47921; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.node_resources_used
    ADD CONSTRAINT "FK_75870a8ed1c14efd1dd4ef47921" FOREIGN KEY (node_id) REFERENCES public.node(id);


--
-- Name: node_resources_free FK_923c4dff43306d0a0f5a98a1aba; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.node_resources_free
    ADD CONSTRAINT "FK_923c4dff43306d0a0f5a98a1aba" FOREIGN KEY (node_id) REFERENCES public.node(id);


--
-- Name: node FK_d224b7b862841f24dd85b556059; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.node
    ADD CONSTRAINT "FK_d224b7b862841f24dd85b556059" FOREIGN KEY (location_id) REFERENCES public.location(id);


--
-- Name: public_config FK_d394b8b9afbb1b1a2346f9743cd; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.public_config
    ADD CONSTRAINT "FK_d394b8b9afbb1b1a2346f9743cd" FOREIGN KEY (node_id) REFERENCES public.node(id);


--
-- Name: node_contract FK_f294cfb50bb7c7b976d86c08fda; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.node_contract
    ADD CONSTRAINT "FK_f294cfb50bb7c7b976d86c08fda" FOREIGN KEY (resources_used_id) REFERENCES public.contract_resources(id);


--
-- Name: node_resources_total FK_fd430c3a2645c8f409f859c2aae; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.node_resources_total
    ADD CONSTRAINT "FK_fd430c3a2645c8f409f859c2aae" FOREIGN KEY (node_id) REFERENCES public.node(id);


--
-- PostgreSQL database dump complete
--
