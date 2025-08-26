--
-- PostgreSQL database dump
--

\restrict kjat48JEwxUStXA4GnmmegyFEUjW5UWBb3apRFcL9rp6CJIGB2HyW9kbVVQuulT

-- Dumped from database version 16.10 (Homebrew)
-- Dumped by pg_dump version 16.10 (Homebrew)

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

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: posts; Type: TABLE; Schema: public; Owner: <username>
--

CREATE TABLE public.posts (
    user_id integer NOT NULL,
    parent_id integer NOT NULL,
    thread_id integer NOT NULL,
    username character varying(30) NOT NULL,
    content character varying(350) NOT NULL,
    created_on date DEFAULT CURRENT_DATE NOT NULL,
    CONSTRAINT non_empty_column CHECK ((length(TRIM(BOTH FROM content)) > 0))
);


ALTER TABLE public.posts OWNER TO <username>;

--
-- Name: posts_thread_id_seq; Type: SEQUENCE; Schema: public; Owner: <username>
--

CREATE SEQUENCE public.posts_thread_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.posts_thread_id_seq OWNER TO <username>;

--
-- Name: posts_thread_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: <username>
--

ALTER SEQUENCE public.posts_thread_id_seq OWNED BY public.posts.thread_id;


--
-- Name: user_account; Type: TABLE; Schema: public; Owner: <username>
--

CREATE TABLE public.user_account (
    id integer NOT NULL,
    username character varying(40) NOT NULL,
    email character varying(30) NOT NULL,
    password bytea NOT NULL,
    followers character varying[]
);


ALTER TABLE public.user_account OWNER TO <username>;

--
-- Name: user_account_id_seq; Type: SEQUENCE; Schema: public; Owner: <username>
--

CREATE SEQUENCE public.user_account_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.user_account_id_seq OWNER TO <username>;

--
-- Name: user_account_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: <username>
--

ALTER SEQUENCE public.user_account_id_seq OWNED BY public.user_account.id;


--
-- Name: user_profile; Type: TABLE; Schema: public; Owner: <username>
--

CREATE TABLE public.user_profile (
    id integer NOT NULL,
    user_id integer NOT NULL,
    fname character varying(20) NOT NULL,
    lname character varying(40) NOT NULL,
    address character varying(80) NOT NULL
);


ALTER TABLE public.user_profile OWNER TO <username>;

--
-- Name: user_profile_id_seq; Type: SEQUENCE; Schema: public; Owner: <username>
--

CREATE SEQUENCE public.user_profile_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.user_profile_id_seq OWNER TO <username>;

--
-- Name: user_profile_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: <username>
--

ALTER SEQUENCE public.user_profile_id_seq OWNED BY public.user_profile.id;


--
-- Name: posts thread_id; Type: DEFAULT; Schema: public; Owner: <username>
--

ALTER TABLE ONLY public.posts ALTER COLUMN thread_id SET DEFAULT nextval('public.posts_thread_id_seq'::regclass);


--
-- Name: user_account id; Type: DEFAULT; Schema: public; Owner: <username>
--

ALTER TABLE ONLY public.user_account ALTER COLUMN id SET DEFAULT nextval('public.user_account_id_seq'::regclass);


--
-- Name: user_profile id; Type: DEFAULT; Schema: public; Owner: <username>
--

ALTER TABLE ONLY public.user_profile ALTER COLUMN id SET DEFAULT nextval('public.user_profile_id_seq'::regclass);


--
-- Data for Name: posts; Type: TABLE DATA; Schema: public; Owner: <username>
--

COPY public.posts (user_id, parent_id, thread_id, username, content, created_on) FROM stdin;
\.


--
-- Data for Name: user_account; Type: TABLE DATA; Schema: public; Owner: <username>
--

COPY public.user_account (id, username, email, password, followers) FROM stdin;
\.


--
-- Data for Name: user_profile; Type: TABLE DATA; Schema: public; Owner: <username>
--

COPY public.user_profile (id, user_id, fname, lname, address) FROM stdin;
\.


--
-- Name: posts_thread_id_seq; Type: SEQUENCE SET; Schema: public; Owner: <username>
--

SELECT pg_catalog.setval('public.posts_thread_id_seq', 1, false);


--
-- Name: user_account_id_seq; Type: SEQUENCE SET; Schema: public; Owner: <username>
--

SELECT pg_catalog.setval('public.user_account_id_seq', 1, false);


--
-- Name: user_profile_id_seq; Type: SEQUENCE SET; Schema: public; Owner: <username>
--

SELECT pg_catalog.setval('public.user_profile_id_seq', 1, false);


--
-- Name: user_account user_account_email_key; Type: CONSTRAINT; Schema: public; Owner: <username>
--

ALTER TABLE ONLY public.user_account
    ADD CONSTRAINT user_account_email_key UNIQUE (email);


--
-- Name: user_account user_account_followers_key; Type: CONSTRAINT; Schema: public; Owner: <username>
--

ALTER TABLE ONLY public.user_account
    ADD CONSTRAINT user_account_followers_key UNIQUE (followers);


--
-- Name: user_account user_account_password_key; Type: CONSTRAINT; Schema: public; Owner: <username>
--

ALTER TABLE ONLY public.user_account
    ADD CONSTRAINT user_account_password_key UNIQUE (password);


--
-- Name: user_account user_account_pkey; Type: CONSTRAINT; Schema: public; Owner: <username>
--

ALTER TABLE ONLY public.user_account
    ADD CONSTRAINT user_account_pkey PRIMARY KEY (id);


--
-- Name: user_account user_account_username_key; Type: CONSTRAINT; Schema: public; Owner: <username>
--

ALTER TABLE ONLY public.user_account
    ADD CONSTRAINT user_account_username_key UNIQUE (username);


--
-- Name: user_profile user_profile_pkey; Type: CONSTRAINT; Schema: public; Owner: <username>
--

ALTER TABLE ONLY public.user_profile
    ADD CONSTRAINT user_profile_pkey PRIMARY KEY (id);


--
-- Name: user_profile user_profile_user_id_fname_lname_address_key; Type: CONSTRAINT; Schema: public; Owner: <username>
--

ALTER TABLE ONLY public.user_profile
    ADD CONSTRAINT user_profile_user_id_fname_lname_address_key UNIQUE (user_id, fname, lname, address);


--
-- Name: user_profile user_profile_user_id_key; Type: CONSTRAINT; Schema: public; Owner: <username>
--

ALTER TABLE ONLY public.user_profile
    ADD CONSTRAINT user_profile_user_id_key UNIQUE (user_id);


--
-- Name: posts fk_user_account_id; Type: FK CONSTRAINT; Schema: public; Owner: <username>
--

ALTER TABLE ONLY public.posts
    ADD CONSTRAINT fk_user_account_id FOREIGN KEY (user_id) REFERENCES public.user_account(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: posts fk_user_account_username; Type: FK CONSTRAINT; Schema: public; Owner: <username>
--

ALTER TABLE ONLY public.posts
    ADD CONSTRAINT fk_user_account_username FOREIGN KEY (username) REFERENCES public.user_account(username) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: user_profile user_profile_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: <username>
--

ALTER TABLE ONLY public.user_profile
    ADD CONSTRAINT user_profile_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.user_account(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- PostgreSQL database dump complete
--

\unrestrict kjat48JEwxUStXA4GnmmegyFEUjW5UWBb3apRFcL9rp6CJIGB2HyW9kbVVQuulT

