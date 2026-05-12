#!/usr/bin/env python3
# -*- coding: utf-8 -*-

import json
import logging
import os
import re
import sys
import time
import uuid
import hashlib
import secrets
from urllib.parse import urlparse

import requests
from Crypto.Cipher import AES, PKCS1_v1_5
from Crypto.PublicKey import RSA

"""
DSM 面板适配版：
- 一个变量组对应一个账号，优先读取 IKUUU_EMAIL / IKUUU_PASSWD。
- Cookie、domain_state.json 等运行状态写入 DSM_STATE_DIR，不再混到脚本源码目录。
- 如需自动发现域名，可在变量组设置 IKUUU_DISCOVERY_URL / IKUUU_MAIN_URL。
- 如只想固定主站，可在变量组设置 IKUUU_BASE_URL。
"""

# =========================
# 基础配置
# =========================

logging.basicConfig(
    level=getattr(logging, os.getenv("LOG_LEVEL", "INFO").upper(), logging.INFO),
    format="%(asctime)s %(levelname)s %(name)s: %(message)s",
    stream=sys.stdout,
)
logger = logging.getLogger("ikuuu")

CAPTCHA_ID = "cc96d05ba8b60f9112f76e18526fcb73"
GT4_BASE   = "https://gcaptcha4.geetest.com"
DEFAULT_IKUUU_BASE_URL = "https://ikuuu.win"
SCRIPT_DIR = os.path.dirname(os.path.abspath(__file__))
STATE_DIR = os.getenv("DSM_STATE_DIR") or os.getenv("DSM_WORKDIR") or SCRIPT_DIR
STATE_DIR = os.path.abspath(STATE_DIR)
os.makedirs(STATE_DIR, exist_ok=True)

DOMAIN_STATE_FILE = os.getenv("IKUUU_DOMAIN_STATE_FILE") or os.path.join(
    STATE_DIR,
    "domain_state.json",
)
ACCOUNTS_FILE = os.getenv("IKUUU_ACCOUNTS_FILE") or os.path.join(
    STATE_DIR,
    "accounts.json",
)
REQUEST_TIMEOUT = 20
DISCOVERY_TIMEOUT_MS = 15000
POW_TIMEOUT_SECONDS = 15
POW_MAX_ATTEMPTS = 200000
CHECKIN_PATH = "/user/checkin"
USER_PATH = "/user"

UA = (
    "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) "
    "AppleWebKit/537.36 (KHTML, like Gecko) "
    "Chrome/147.0.0.0 Safari/537.36"
)

COMMON_HEADERS = {
    "user-agent": UA,
    "accept-language": "zh-CN,zh;q=0.9,en-US;q=0.8",
}

RSA_E_HEX = "10001"
RSA_N_HEX = (
    "c1e3934d1614465b33053e7f48ee4ec87b14b95ef88947713d25eecbff7e74c7"
    "977d02dc1d9451f79dd5d1c10c29acb6a9b4d6fb7d0a0279b6719e1772565f09"
    "af627715919221aef91899cae08c0d686d748b20a3603be2318ca6bc2b597065"
    "92a9219d0bf05c9f65023a21d2330807252ae0066d59ceefa5f2748ea80bab81"
)

# =========================
# 工具函数
# =========================

def get_env(name, default=""):
    return os.getenv(name, default).strip()

def strip_wrapping_quotes(value):
    value = (value or "").strip()
    if len(value) >= 2:
        if (value[0] == "'" and value[-1] == "'") or (value[0] == '"' and value[-1] == '"'):
            value = value[1:-1].strip()
    return value

def canonical_email_id(email):
    return strip_wrapping_quotes(email).lower()

def normalize_account_item(item, index):
    if not isinstance(item, dict):
        raise RuntimeError(f"accounts.json 第 {index + 1} 项不是对象。")

    email = strip_wrapping_quotes(item.get("email", item.get("IKUUU_EMAIL", "")))
    passwd = strip_wrapping_quotes(item.get("passwd", item.get("IKUUU_PASSWD", "")))
    if not email or not passwd:
        raise RuntimeError(
            f"accounts.json 第 {index + 1} 项缺少 email/passwd。"
        )
    return {"email": email, "passwd": passwd}

def load_accounts_from_json():
    if not os.path.isfile(ACCOUNTS_FILE):
        return []

    try:
        with open(ACCOUNTS_FILE, "r", encoding="utf-8-sig") as f:
            data = json.load(f)
    except Exception as e:
        raise RuntimeError(f"读取 accounts.json 失败: {e}")

    if isinstance(data, dict):
        raw_accounts = data.get("accounts", [])
    elif isinstance(data, list):
        raw_accounts = data
    else:
        raise RuntimeError("accounts.json 格式无效，必须是数组或包含 accounts 数组的对象。")

    if not isinstance(raw_accounts, list):
        raise RuntimeError("accounts.json 中的 accounts 必须是数组。")
    if not raw_accounts:
        raise RuntimeError("accounts.json 已存在，但 accounts 为空。")

    accounts = []
    seen = set()
    for index, item in enumerate(raw_accounts):
        account = normalize_account_item(item, index)
        email = account["email"]
        email_id = canonical_email_id(email)
        if email_id in seen:
            raise RuntimeError(f"accounts.json 中存在重复邮箱: {email}")
        seen.add(email_id)
        accounts.append(account)

    if accounts:
        logger.info("从 accounts.json 加载 %s 个账号", len(accounts))
    return accounts

def load_accounts():
    email = strip_wrapping_quotes(get_env("IKUUU_EMAIL"))
    passwd = strip_wrapping_quotes(get_env("IKUUU_PASSWD"))
    if email and passwd:
        return [{"email": email, "passwd": passwd}]

    if os.path.isfile(ACCOUNTS_FILE):
        return load_accounts_from_json()

    if not email or not passwd:
        raise RuntimeError(
            "未找到账号配置。请在 DSM 变量组设置 IKUUU_EMAIL / IKUUU_PASSWD，"
            f"或把 accounts.json 放到状态目录: {ACCOUNTS_FILE}"
        )

def mask_email(email):
    email = canonical_email_id(email)
    if not email or "@" not in email:
        return email or "(unknown)"

    local, domain = email.split("@", 1)
    if len(local) <= 2:
        masked_local = local[0] + "*" * max(len(local) - 1, 0)
    else:
        masked_local = local[0] + "*" * (len(local) - 2) + local[-1]
    return f"{masked_local}@{domain}"


def sanitize_filename(value):
    cleaned = re.sub(r'[<>:"/\\|?*]', "_", (value or "").strip())
    cleaned = cleaned.rstrip(". ")
    if cleaned:
        return cleaned
    return f"account_{hashlib.md5((value or '').encode('utf-8')).hexdigest()[:8]}"

def account_state_key(email):
    env_group_id = get_env("DSM_ENV_GROUP_ID")
    if env_group_id:
        return f"{env_group_id}_{canonical_email_id(email)}"

    env_group_name = get_env("DSM_ENV_GROUP_NAME")
    if env_group_name:
        return f"{env_group_name}_{canonical_email_id(email)}"

    return canonical_email_id(email)

def get_cookie_file(email):
    safe_key = sanitize_filename(account_state_key(email))
    return os.path.join(STATE_DIR, f"IKUUU_COOKIE_{safe_key}.txt")

def tg_targets():
    allow_from = get_env("ALLOW_FROM")
    if not allow_from:
        return []

    targets = []
    seen = set()
    for item in allow_from.split(","):
        chat_id = item.strip()
        if chat_id and chat_id not in seen:
            seen.add(chat_id)
            targets.append(chat_id)
    return targets

def tg_notify(text):
    token = get_env("TG_BOT_TOKEN")
    targets = tg_targets()
    if not token or not targets:
        return

    proxy = get_env("PROXY")
    proxies = {"http": proxy, "https": proxy} if proxy else None
    api = f"https://api.telegram.org/bot{token}/sendMessage"

    for chat_id in targets:
        try:
            resp = requests.post(
                api,
                json={
                    "chat_id": chat_id,
                    "text": text,
                    "disable_web_page_preview": True,
                },
                headers={"user-agent": UA},
                timeout=REQUEST_TIMEOUT,
                proxies=proxies,
            )
            if resp.status_code != 200:
                logger.warning(
                    "Telegram 通知失败: chat_id=%s, status=%s, body=%s",
                    chat_id,
                    resp.status_code,
                    resp.text[:300],
                )
        except requests.RequestException as e:
            logger.warning("Telegram 通知异常: chat_id=%s, error=%s", chat_id, e)

def normalize_base_url(url):
    url = (url or "").strip()
    if len(url) >= 2:
        if (url[0] == "'" and url[-1] == "'") or (url[0] == '"' and url[-1] == '"'):
            url = url[1:-1].strip()
    if not url:
        return ""
    if not re.match(r"^https?://", url, flags=re.I):
        url = "https://" + url
    return url.rstrip("/")

def url_origin(url):
    base = normalize_base_url(url)
    if not base:
        return ""
    parsed = urlparse(base)
    if not parsed.scheme or not parsed.netloc:
        return base
    return f"{parsed.scheme}://{parsed.netloc}"

def get_cookie(email, default=""):
    cookie_file = get_cookie_file(email)
    try:
        with open(cookie_file, "r", encoding="utf-8") as f:
            return f.read().strip()
    except FileNotFoundError:
        return default

def save_cookie(email, cookie):
    cookie_file = get_cookie_file(email)
    with open(cookie_file, "w", encoding="utf-8") as f:
        f.write(cookie.strip())

def load_domain_state():
    env_discovery_url = normalize_base_url(get_env("IKUUU_DISCOVERY_URL"))
    env_main_url = normalize_base_url(get_env("IKUUU_MAIN_URL"))
    if env_discovery_url and env_main_url:
        save_domain_state(env_discovery_url, env_main_url)
        return {
            "discovery_url": env_discovery_url,
            "main_url": env_main_url,
        }

    try:
        with open(DOMAIN_STATE_FILE, "r", encoding="utf-8") as f:
            data = json.load(f)
    except FileNotFoundError:
        raise RuntimeError(
            "缺少域名配置。请在 DSM 变量组设置 IKUUU_BASE_URL，"
            "或同时设置 IKUUU_DISCOVERY_URL / IKUUU_MAIN_URL，"
            f"也可以把 domain_state.json 放到状态目录: {DOMAIN_STATE_FILE}。"
            '格式: {"discovery_url": "https://你的域名获取页", "main_url": "https://你当前主站"}'
        )
    except Exception as e:
        raise RuntimeError(f"读取 domain_state.json 失败: {e}")

    discovery_url = normalize_base_url(data.get("discovery_url"))
    main_url = normalize_base_url(data.get("main_url"))
    if not discovery_url or not main_url:
        raise RuntimeError(
            "domain_state.json 格式无效，必须同时包含 discovery_url 和 main_url。"
        )

    return {
        "discovery_url": discovery_url,
        "main_url": main_url,
    }

def save_domain_state(discovery_url, main_url):
    payload = {
        "discovery_url": normalize_base_url(discovery_url),
        "main_url": normalize_base_url(main_url),
    }
    with open(DOMAIN_STATE_FILE, "w", encoding="utf-8") as f:
        json.dump(payload, f, ensure_ascii=False, indent=2)

def persist_manual_base_url(resolved_base):
    try:
        state = load_domain_state()
        discovery_url = state["discovery_url"]
    except Exception:
        discovery_url = resolved_base

    try:
        save_domain_state(discovery_url, resolved_base)
    except Exception as e:
        logger.warning("持久化手动指定域名失败: %s", e)

def _parse(text):
    text = text.strip()
    if text and text[0] != "{":
        import re
        m = re.search(r"\((.*)\)\s*;?\s*$", text, re.S)
        if m:
            return json.loads(m.group(1))
    return json.loads(text)


def _pad(data):
    pad = 16 - len(data) % 16
    return data + bytes([pad]) * pad

def _aes_encrypt(data, key):
    cipher = AES.new(key.encode(), AES.MODE_CBC, iv=b"0000000000000000")
    return cipher.encrypt(_pad(data.encode())).hex()

def _rsa_encrypt(key):
    pub = RSA.construct((int(RSA_N_HEX, 16), int(RSA_E_HEX, 16)))
    cipher = PKCS1_v1_5.new(pub)
    return cipher.encrypt(key.encode()).hex()

def _rand_key():
    return "".join(f"{secrets.randbelow(65536):04x}" for _ in range(4))

def _hash(msg, algo):
    h = hashlib.new(algo)
    h.update(msg.encode())
    return h.hexdigest()

# =========================
# GT4 核心
# =========================

def solve_pow(lot, captcha_id, detail):
    prefix = f"{detail['version']}|{detail['bits']}|{detail['hashfunc']}|{detail['datetime']}|{captcha_id}|{lot}||"
    bits = detail["bits"]
    deadline = time.monotonic() + POW_TIMEOUT_SECONDS
    attempts = 0

    while True:
        attempts += 1
        if attempts > POW_MAX_ATTEMPTS or time.monotonic() >= deadline:
            raise RuntimeError(
                f"GT4 PoW 求解超时: bits={bits}, attempts={attempts}, timeout={POW_TIMEOUT_SECONDS}s"
            )

        suffix = secrets.token_hex(8)
        msg = prefix + suffix
        sign = _hash(msg, detail["hashfunc"])
        binary = bin(int(sign, 16))[2:].zfill(len(sign) * 4)
        if binary.startswith("0" * bits):
            return msg, sign

def _js_path_fields(lot):
    key = lot[3:6] + lot[9:12]
    val = lot[7:13]
    return {"1a8R": "daC2", key: val}

def generate_w(data):
    key = _rand_key()
    plain = json.dumps(data, separators=(",", ":"))
    return _aes_encrypt(plain, key) + _rsa_encrypt(key)

def solve_captcha(session):
    cb = f"geetest_{int(time.time()*1000)}"

    # load
    load = session.get(
        f"{GT4_BASE}/load",
        params={
            "callback": cb,
            "captcha_id": CAPTCHA_ID,
            "challenge": str(uuid.uuid4()),
            "client_type": "web",
            "risk_type": "ai",
        },
        headers=COMMON_HEADERS,
        timeout=REQUEST_TIMEOUT,
    )
    logger.debug(load.status_code)
    logger.debug(load.text)
    ld = _parse(load.text)["data"]

    # pow
    pow_msg, pow_sign = solve_pow(ld["lot_number"], CAPTCHA_ID, ld["pow_detail"])

    lot = ld["lot_number"]
    data = {
        "device_id": "",
        "lot_number": lot,
        "pow_msg": pow_msg,
        "pow_sign": pow_sign,
        "geetest": "captcha",
        "lang": "zh",
        "ep": "123",
        "biht": "1426265548",
        **_js_path_fields(lot),
        "em": {"ph": 0, "cp": 0, "ek": "11", "wd": 1, "nt": 0, "si": 0, "sc": 0},
    }
    if ld.get("guard"):
        data["gee_guard"] = {"roe": {"aup": "3", "sep": "3", "egp": "3", "auh": "3",
                                     "rew": "3", "snh": "3", "res": "3", "cdc": "3"}}

    w = generate_w(data)

    # verify
    cb2 = f"geetest_{int(time.time()*1000)}"
    verify = session.get(
        f"{GT4_BASE}/verify",
        params={
            "callback": cb2,
            "captcha_id": CAPTCHA_ID,
            "lot_number": ld["lot_number"],
            "payload": ld["payload"],
            "process_token": ld["process_token"],
            "payload_protocol": ld["payload_protocol"],
            "w": w,
        },
        headers=COMMON_HEADERS,
        timeout=REQUEST_TIMEOUT,
    )

    vd = _parse(verify.text)["data"]["seccode"]

    return {
        "lot_number": ld["lot_number"],
        "captcha_output": vd["captcha_output"],
        "pass_token": vd["pass_token"],
        "gen_time": vd["gen_time"],
    }

# =========================
# 登录 + 签到
# =========================

def build_login_headers(base_url):
    return {
        "accept": "application/json, text/javascript, */*; q=0.01",
        "accept-language": "zh-CN,zh;q=0.9,en-US;q=0.8,en;q=0.7",
        "cache-control": "no-cache",
        "content-type": "application/x-www-form-urlencoded; charset=UTF-8",
        "origin": base_url,
        "pragma": "no-cache",
        "priority": "u=1, i",
        "referer": f"{base_url}/auth/login",
        "sec-ch-ua": '"Google Chrome";v="147", "Not.A/Brand";v="8", "Chromium";v="147"',
        "sec-ch-ua-mobile": "?0",
        "sec-ch-ua-platform": '"macOS"',
        "sec-fetch-dest": "empty",
        "sec-fetch-mode": "cors",
        "sec-fetch-site": "same-origin",
        "user-agent": UA,
        "x-requested-with": "XMLHttpRequest",
    }

def extract_candidate_domains_via_playwright(source_url):
    try:
        from playwright.sync_api import sync_playwright
    except Exception as e:
        raise RuntimeError(
            "Playwright 不可用，请先安装: pip install playwright && playwright install chromium"
        ) from e

    candidates = []
    seen = set()

    def add_candidate(raw):
        base = url_origin(raw)
        if not base:
            return

        host = urlparse(base).netloc.lower()
        if "ikuuu." not in host:
            return

        if base not in seen:
            seen.add(base)
            candidates.append(base)

    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        page = browser.new_page()
        try:
            page.goto(source_url, wait_until="domcontentloaded", timeout=DISCOVERY_TIMEOUT_MS)
            page.wait_for_timeout(1500)
            hrefs = page.eval_on_selector_all(
                "#domain-list a[href]",
                "els => els.map(e => e.href)",
            )
            for href in hrefs:
                parsed = urlparse(href)
                if parsed.scheme and parsed.netloc:
                    add_candidate(f"{parsed.scheme}://{parsed.netloc}")
                else:
                    add_candidate(href)
        finally:
            browser.close()

    return candidates

def looks_like_login_page(resp):
    body = resp.text.lower()
    markers = (
        'name="email"',
        'name="passwd"',
        "remember_me",
        "/auth/login",
        "type=\"password\"",
    )
    return sum(1 for marker in markers if marker in body) >= 2

def probe_base_url(base_url):
    try:
        login_resp = requests.get(
            f"{base_url}/auth/login",
            headers={"user-agent": UA},
            timeout=10,
            allow_redirects=True,
        )
        final_base = url_origin(login_resp.url)
        if login_resp.status_code == 200 and looks_like_login_page(login_resp):
            return True, final_base, "login_page"
    except requests.RequestException as e:
        return False, base_url, str(e)

    try:
        user_resp = requests.get(
            f"{base_url}{USER_PATH}",
            headers={"user-agent": UA},
            timeout=10,
            allow_redirects=True,
        )
        final_base = url_origin(user_resp.url)
        final_path = urlparse(user_resp.url).path.lower()
        if user_resp.status_code == 200 and "/auth/login" in final_path and looks_like_login_page(user_resp):
            return True, final_base, "user_redirect_login"
        if user_resp.status_code in (200, 403) and final_path.startswith(USER_PATH):
            return True, final_base, "user_page"
    except requests.RequestException as e:
        return False, base_url, str(e)

    return False, base_url, "missing_login_markers"

def discover_main_from_source(source_url):
    candidates = extract_candidate_domains_via_playwright(source_url)
    logger.info(
        "发现源候选主站 [%s]: %s",
        source_url,
        ", ".join(candidates) if candidates else "(空)",
    )

    if not candidates:
        raise RuntimeError(f"{source_url} 没有返回任何主站候选")

    return candidates[0]

def resolve_base_url():
    manual_url = normalize_base_url(get_env("IKUUU_BASE_URL", DEFAULT_IKUUU_BASE_URL))
    if manual_url:
        ok, resolved_base, reason = probe_base_url(manual_url)
        if ok:
            persist_manual_base_url(resolved_base)
            if resolved_base != manual_url:
                logger.info("手动指定域名重定向到: %s -> %s (%s)", manual_url, resolved_base, reason)
            return resolved_base, "manual"
        logger.warning("手动指定域名不可用: %s", manual_url)

    state = load_domain_state()
    discovery_url = state["discovery_url"]
    main_url = state["main_url"]
    logger.info("当前域名状态: 获取页=%s, 主站=%s", discovery_url, main_url)

    discovery_error = None
    try:
        discovered_main = discover_main_from_source(discovery_url)
        next_discovery_url = discovery_url

        if discovered_main != main_url:
            try:
                main_discovered = discover_main_from_source(main_url)
                if normalize_base_url(main_discovered) == normalize_base_url(discovered_main):
                    next_discovery_url = main_url
                    logger.info("旧主站可继续承担获取页，切换获取页为: %s", next_discovery_url)
            except Exception as e:
                logger.info("旧主站暂不能承担获取页，继续使用原获取页 %s: %s", discovery_url, e)

        save_domain_state(next_discovery_url, discovered_main)
        return discovered_main, "auto"
    except Exception as e:
        discovery_error = str(e)
        logger.warning("域名获取页失效 [%s]: %s", discovery_url, discovery_error)

    try:
        discovered_main = discover_main_from_source(main_url)
        save_domain_state(main_url, discovered_main)
        logger.info("主站兼任获取页成功，切换获取页为: %s", main_url)
        return discovered_main, "auto"
    except Exception as main_error:
        raise RuntimeError(
            f"域名获取页无法获取主站: {discovery_url} -> {discovery_error}; {main_url} -> {main_error}"
        )

def do_login(base_url, email, passwd):
    session = requests.Session()
    session.headers.update(COMMON_HEADERS)
    captcha = solve_captcha(session)

    session.headers.update(build_login_headers(base_url))
    resp = session.post(
        f"{base_url}/auth/login",
        data={
            "host":   urlparse(base_url).netloc,
            "email":  email,
            "passwd": passwd,
            "code":   "",
            "captcha_result[lot_number]":     captcha["lot_number"],
            "captcha_result[captcha_output]": captcha["captcha_output"],
            "captcha_result[pass_token]":     captcha["pass_token"],
            "captcha_result[gen_time]":       captcha["gen_time"],
            "remember_me":  "on",
            "pageLoadedAt": str(int(time.time() * 1000)),
        },
        timeout=REQUEST_TIMEOUT,
    )
    logger.debug(resp.text)

    if resp.status_code != 200:
        raise RuntimeError(f"登录请求失败: HTTP {resp.status_code}, body={resp.text[:300]}")

    try:
        data = resp.json()
    except ValueError as e:
        raise RuntimeError(f"登录返回不是 JSON: {resp.text[:300]}") from e

    if data.get("ret") != 1:
        raise RuntimeError(f"登录失败: {format_result_text(data)}")

    cookies = dict(session.cookies)
    if not cookies:
        raise RuntimeError("登录成功但未获取到 cookie")
    return cookies

def do_checkin(base_url, cookie):
    try:
        resp = requests.post(
            f"{base_url}{CHECKIN_PATH}",
            headers={
                "cookie": cookie,
                "user-agent": UA,
                "x-requested-with": "XMLHttpRequest",
                "origin": base_url,
                "referer": f"{base_url}{USER_PATH}",
                "accept": "application/json, text/javascript, */*; q=0.01",
            },
            timeout=REQUEST_TIMEOUT,
            allow_redirects=False,
        )
    except requests.RequestException as e:
        return False, {"error": f"签到请求异常: {e}"}

    if resp.status_code in (301, 302, 303, 307, 308):
        location = resp.headers.get("location", "")
        if "/auth/login" in location.lower():
            return False, {
                "status_code": resp.status_code,
                "redirect_to": location,
                "error": f"签到请求被重定向到登录页: {location}",
            }
        return False, {
            "status_code": resp.status_code,
            "redirect_to": location,
            "error": f"签到请求被重定向: HTTP {resp.status_code}, location={location}",
        }

    if resp.status_code != 200:
        return False, {
            "status_code": resp.status_code,
            "error": f"签到请求失败: HTTP {resp.status_code}, body={resp.text[:300]}",
        }

    if looks_like_login_page(resp):
        return False, {"error": "签到请求返回登录页，cookie 可能已失效"}

    try:
        data = resp.json()
    except ValueError:
        return False, {"error": f"签到返回不是 JSON: {resp.text[:300]}"}

    if data.get("ret") == 1:
        return True, data

    if not any(key in data for key in ("msg", "error", "ret")):
        return False, {"error": f"签到返回缺少关键字段: {json.dumps(data, ensure_ascii=False)[:300]}"}

    return False, data


def should_relogin(result):
    if not isinstance(result, dict):
        return True

    if result.get("status_code") == 403:
        return True

    msg = str(result.get("msg", "")).strip()
    error = str(result.get("error", "")).strip()
    text = f"{msg} {error}".lower()

    if "已经签到过" in msg:
        return False

    relogin_markers = (
        "未登录",
        "请先登录",
        "登录",
        "cookie",
        "auth",
        "302",
        "csrf",
        "forbidden",
        "redirect",
    )
    return any(marker.lower() in text for marker in relogin_markers)

def is_already_checked_in(result):
    if not isinstance(result, dict):
        return False
    msg = str(result.get("msg", "")).strip()
    return "已经签到过" in msg

def format_result_text(result):
    if isinstance(result, dict):
        msg = str(result.get("msg", "")).strip()
        error = str(result.get("error", "")).strip()
        if msg:
            return msg
        if error:
            return error
        return json.dumps(result, ensure_ascii=False)
    return str(result)

def run_account(base_url, source, email, passwd):
    masked_email = mask_email(email)
    logger.info("开始处理账号: %s", masked_email)
    cookie = get_cookie(email)

    try:
        if not cookie:
            logger.info("账号 %s 的 cookie 文件不存在或为空，尝试登录", masked_email)
            cookies = do_login(base_url, email, passwd)
            cookie = "; ".join(f"{k}={v}" for k, v in cookies.items())
            save_cookie(email, cookie)

        ok, result = do_checkin(base_url, cookie)

        if not ok and should_relogin(result):
            logger.info("账号 %s 的 cookie 失效，尝试重新登录", masked_email)
            cookies = do_login(base_url, email, passwd)
            cookie = "; ".join(f"{k}={v}" for k, v in cookies.items())
            save_cookie(email, cookie)
            ok, result = do_checkin(base_url, cookie)
    except Exception as e:
        result = {"error": str(e)}
        ok = False

    if ok:
        status = "success"
        message = (
            f"✅ iKuuu 签到成功\n"
            f"账号: {masked_email}\n"
            f"域名: {base_url}\n"
            f"来源: {source}\n"
            f"结果: {format_result_text(result)}"
        )
    elif is_already_checked_in(result):
        status = "already"
        message = (
            f"✅ iKuuu 今日已签到\n"
            f"账号: {masked_email}\n"
            f"域名: {base_url}\n"
            f"来源: {source}\n"
            f"结果: {format_result_text(result)}"
        )
    else:
        status = "failed"
        message = (
            f"❌ iKuuu 签到失败\n"
            f"账号: {masked_email}\n"
            f"域名: {base_url}\n"
            f"来源: {source}\n"
            f"原因: {format_result_text(result)}"
        )

    logger.info(message.replace("\n", " | "))
    return {
        "email": email,
        "status": status,
        "result": result,
        "message": message,
        "cookie_file": get_cookie_file(email),
    }

# =========================
# 主流程
# =========================

def main():
    logger.info("状态目录: %s", STATE_DIR)

    try:
        accounts = load_accounts()
    except Exception as e:
        result = {"error": f"加载账号失败: {e}"}
        print(json.dumps(result, ensure_ascii=False))
        tg_notify(f"❌ iKuuu 签到失败\n原因: {result['error']}")
        return

    try:
        base_url, source = resolve_base_url()
    except Exception as e:
        result = {"error": f"获取域名失败: {e}"}
        print(json.dumps(result, ensure_ascii=False))
        tg_notify(f"❌ iKuuu 签到失败\n原因: {result['error']}")
        return

    logger.info("使用域名: %s (来源: %s)", base_url, source)
    results = []
    for account in accounts:
        results.append(run_account(base_url, source, account["email"], account["passwd"]))

    tg_notify("\n\n".join(item["message"] for item in results))

    if len(results) == 1:
        print(json.dumps(results[0]["result"], ensure_ascii=False))
    else:
        print(json.dumps([
            {
                "email": mask_email(item["email"]),
                "status": item["status"],
                "result": item["result"],
                "cookie_file": item["cookie_file"],
            }
            for item in results
        ], ensure_ascii=False, indent=2))

if __name__ == "__main__":
    main()
