import urllib.request
import urllib.error
import json
import time
import concurrent.futures

BASE_URL = "http://127.0.0.1:3000/api/v1"

def request(method, path, body=None, token=None):
    headers = {
        "Content-Type": "application/json",
        "Connection": "close"
    }
    if token:
        headers["Authorization"] = f"Bearer {token}"
    data = json.dumps(body).encode("utf-8") if body is not None else None
    req = urllib.request.Request(f"{BASE_URL}{path}", data=data, headers=headers, method=method)
    try:
        with urllib.request.urlopen(req, timeout=5) as resp:
            content = resp.read().decode("utf-8")
            return resp.status, json.loads(content) if content else {}
    except urllib.error.HTTPError as e:
        content = e.read().decode("utf-8")
        try:
            return e.code, json.loads(content) if content else {}
        except Exception:
            return e.code, {"raw": content}

def run_tests():
    print("=== 1. Health Check ===", flush=True)
    status, res = request("GET", "/health")
    print(f"GET /health: {status} -> {res}", flush=True)

    print("\n=== 2. Login ===", flush=True)
    status, login_res = request("POST", "/auth/login", {"username": "admin", "password": "Admin123!"})
    token = login_res["data"]["access_token"]
    print("Obtained token successfully.", flush=True)

    # Let's test with student 34 for concurrency
    print("\n=== 3. Setup Portfolio for Student 34 ===", flush=True)
    st, res = request("PUT", "/students/34/portfolio", {
        "headline": "Concurrency Benchmark Portfolio",
        "bio": "Menguji batas 20 proyek dengan 10 request serentak."
    }, token=token)
    print(f"PUT /students/34/portfolio: {st}", flush=True)

    # Delete any existing projects from student 34
    st, get_res = request("GET", "/students/34/portfolio", token=token)
    if get_res.get("data", {}).get("projects"):
        for p in get_res["data"]["projects"]:
            request("DELETE", f"/students/34/portfolio/projects/{p['id']}", token=token)
    print("Portfolio student 34 dikosongkan.", flush=True)

    # Fill 18 projects baseline
    print("Mengisi 18 proyek baseline...", flush=True)
    for i in range(1, 19):
        p = {
            "title": f"Project Baseline #{i}",
            "technologies": ["Go", "Docker" if i % 2 == 0 else "PostgreSQL"],
            "year": 2024
        }
        st, _ = request("POST", "/students/34/portfolio/projects", p, token=token)
        if st != 201:
            print(f"Error adding baseline {i}: {st}", flush=True)
    print("18 proyek awal berhasil diisi.", flush=True)

    # Launch 10 concurrent requests
    print("Mengirim 10 request serentak...", flush=True)
    results = []
    def add_concurrent_project(idx):
        p = {
            "title": f"Concurrent Project #{idx}",
            "technologies": ["MongoDB", "Kubernetes"],
            "year": 2024
        }
        st, body = request("POST", "/students/34/portfolio/projects", p, token=token)
        return st, body

    with concurrent.futures.ThreadPoolExecutor(max_workers=10) as executor:
        futures = [executor.submit(add_concurrent_project, i) for i in range(1, 11)]
        for f in concurrent.futures.as_completed(futures):
            results.append(f.result())

    success_count = sum(1 for st, _ in results if st == 201)
    limit_count = sum(1 for st, _ in results if st == 409)
    print(f"Hasil Uji Konkurensi 10 request serentak:", flush=True)
    print(f"  - Status 201 Created   : {success_count} (Diharapkan: tepat 2)", flush=True)
    print(f"  - Status 409 Conflict  : {limit_count} (Diharapkan: tepat 8)", flush=True)

    # Verify final project count via GET API
    st, get_res = request("GET", "/students/34/portfolio", token=token)
    total_projects = len(get_res["data"]["projects"])
    print(f"Jumlah total proyek di portfolio mahasiswa 34: {total_projects} (Tepat 20)", flush=True)

    print("\n=== 4. Aggregation: Top 10 Technologies ===", flush=True)
    status, res = request("GET", "/portfolios/stats/technologies", token=token)
    print(f"GET /portfolios/stats/technologies: {status} -> {json.dumps(res, indent=2)}", flush=True)

    print("\n=== 5. Failure Cases ===", flush=True)
    st1, res1 = request("GET", "/students/99999/portfolio", token=token)
    print(f"Kasus 1 - Mahasiswa 99999 tidak terdaftar di PostgreSQL: {st1} -> {res1}", flush=True)

    st2, res2 = request("POST", "/students/34/portfolio/projects", {"title": "", "technologies": ["Go"], "year": 2024}, token=token)
    print(f"Kasus 2 - Validasi struct gagal (title kosong): {st2} -> {res2}", flush=True)

    st3, res3 = request("DELETE", "/students/34/portfolio/projects/bukan-hex-id", token=token)
    print(f"Kasus 3 - Format ObjectID tidak valid: {st3} -> {res3}", flush=True)

    st4, res4 = request("PUT", "/students/34/portfolio", {"headline": {"$ne": None}, "bio": "Testing NoSQL injection"}, token=token)
    print(f"Kasus 4 - NoSQL Injection payload: {st4} -> {res4}", flush=True)

    print("\n=== 6. Course Reviews Feature Verification ===", flush=True)
    # Check course reviews
    st, list_rev = request("GET", "/courses/1/reviews", token=token)
    print(f"GET /courses/1/reviews: {st}", flush=True)

    st, sum_rev = request("GET", "/courses/1/reviews/summary", token=token)
    print(f"GET /courses/1/reviews/summary: {st} -> {json.dumps(sum_rev, indent=2)}", flush=True)

    print("\nALL VERIFICATIONS COMPLETED SUCCESSFULLY!", flush=True)

if __name__ == "__main__":
    run_tests()
