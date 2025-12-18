// node_modules/@noble/ed25519/index.js
/*! noble-ed25519 - MIT License (c) 2019 Paul Miller (paulmillr.com) */
var ed25519_CURVE = {
  p: 0x7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffedn,
  n: 0x1000000000000000000000000000000014def9dea2f79cd65812631a5cf5d3edn,
  h: 8n,
  a: 0x7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffecn,
  d: 0x52036cee2b6ffe738cc740797779e89800700a4d4141d8ab75eb4dca135978a3n,
  Gx: 0x216936d3cd6e53fec0a4e231fdd6dc5c692cc7609525a7b2c9562d608f25d51an,
  Gy: 0x6666666666666666666666666666666666666666666666666666666666666658n
};
var { p: P, n: N, Gx, Gy, a: _a, d: _d } = ed25519_CURVE;
var h = 8n;
var L = 32;
var L2 = 64;
var err = (m = "") => {
  throw new Error(m);
};
var isBig = (n) => typeof n === "bigint";
var isStr = (s) => typeof s === "string";
var isBytes = (a) => a instanceof Uint8Array || ArrayBuffer.isView(a) && a.constructor.name === "Uint8Array";
var abytes = (a, l) => !isBytes(a) || typeof l === "number" && l > 0 && a.length !== l ? err("Uint8Array expected") : a;
var u8n = (len) => new Uint8Array(len);
var u8fr = (buf) => Uint8Array.from(buf);
var padh = (n, pad) => n.toString(16).padStart(pad, "0");
var bytesToHex = (b) => Array.from(abytes(b)).map((e) => padh(e, 2)).join("");
var C = { _0: 48, _9: 57, A: 65, F: 70, a: 97, f: 102 };
var _ch = (ch) => {
  if (ch >= C._0 && ch <= C._9)
    return ch - C._0;
  if (ch >= C.A && ch <= C.F)
    return ch - (C.A - 10);
  if (ch >= C.a && ch <= C.f)
    return ch - (C.a - 10);
  return;
};
var hexToBytes = (hex) => {
  const e = "hex invalid";
  if (!isStr(hex))
    return err(e);
  const hl = hex.length;
  const al = hl / 2;
  if (hl % 2)
    return err(e);
  const array = u8n(al);
  for (let ai = 0, hi = 0;ai < al; ai++, hi += 2) {
    const n1 = _ch(hex.charCodeAt(hi));
    const n2 = _ch(hex.charCodeAt(hi + 1));
    if (n1 === undefined || n2 === undefined)
      return err(e);
    array[ai] = n1 * 16 + n2;
  }
  return array;
};
var toU8 = (a, len) => abytes(isStr(a) ? hexToBytes(a) : u8fr(abytes(a)), len);
var cr = () => globalThis?.crypto;
var subtle = () => cr()?.subtle ?? err("crypto.subtle must be defined");
var concatBytes = (...arrs) => {
  const r = u8n(arrs.reduce((sum, a) => sum + abytes(a).length, 0));
  let pad = 0;
  arrs.forEach((a) => {
    r.set(a, pad);
    pad += a.length;
  });
  return r;
};
var randomBytes = (len = L) => {
  const c = cr();
  return c.getRandomValues(u8n(len));
};
var big = BigInt;
var arange = (n, min, max, msg = "bad number: out of range") => isBig(n) && min <= n && n < max ? n : err(msg);
var M = (a, b = P) => {
  const r = a % b;
  return r >= 0n ? r : b + r;
};
var modN = (a) => M(a, N);
var invert = (num, md) => {
  if (num === 0n || md <= 0n)
    err("no inverse n=" + num + " mod=" + md);
  let a = M(num, md), b = md, x = 0n, y = 1n, u = 1n, v = 0n;
  while (a !== 0n) {
    const q = b / a, r = b % a;
    const m = x - u * q, n = y - v * q;
    b = a, a = r, x = u, y = v, u = m, v = n;
  }
  return b === 1n ? M(x, md) : err("no inverse");
};
var callHash = (name) => {
  const fn = etc[name];
  if (typeof fn !== "function")
    err("hashes." + name + " not set");
  return fn;
};
var apoint = (p) => p instanceof Point ? p : err("Point expected");
var B256 = 2n ** 256n;

class Point {
  static BASE;
  static ZERO;
  ex;
  ey;
  ez;
  et;
  constructor(ex, ey, ez, et) {
    const max = B256;
    this.ex = arange(ex, 0n, max);
    this.ey = arange(ey, 0n, max);
    this.ez = arange(ez, 1n, max);
    this.et = arange(et, 0n, max);
    Object.freeze(this);
  }
  static fromAffine(p) {
    return new Point(p.x, p.y, 1n, M(p.x * p.y));
  }
  static fromBytes(hex, zip215 = false) {
    const d = _d;
    const normed = u8fr(abytes(hex, L));
    const lastByte = hex[31];
    normed[31] = lastByte & ~128;
    const y = bytesToNumLE(normed);
    const max = zip215 ? B256 : P;
    arange(y, 0n, max);
    const y2 = M(y * y);
    const u = M(y2 - 1n);
    const v = M(d * y2 + 1n);
    let { isValid, value: x } = uvRatio(u, v);
    if (!isValid)
      err("bad point: y not sqrt");
    const isXOdd = (x & 1n) === 1n;
    const isLastByteOdd = (lastByte & 128) !== 0;
    if (!zip215 && x === 0n && isLastByteOdd)
      err("bad point: x==0, isLastByteOdd");
    if (isLastByteOdd !== isXOdd)
      x = M(-x);
    return new Point(x, y, 1n, M(x * y));
  }
  assertValidity() {
    const a = _a;
    const d = _d;
    const p = this;
    if (p.is0())
      throw new Error("bad point: ZERO");
    const { ex: X, ey: Y, ez: Z, et: T } = p;
    const X2 = M(X * X);
    const Y2 = M(Y * Y);
    const Z2 = M(Z * Z);
    const Z4 = M(Z2 * Z2);
    const aX2 = M(X2 * a);
    const left = M(Z2 * M(aX2 + Y2));
    const right = M(Z4 + M(d * M(X2 * Y2)));
    if (left !== right)
      throw new Error("bad point: equation left != right (1)");
    const XY = M(X * Y);
    const ZT = M(Z * T);
    if (XY !== ZT)
      throw new Error("bad point: equation left != right (2)");
    return this;
  }
  equals(other) {
    const { ex: X1, ey: Y1, ez: Z1 } = this;
    const { ex: X2, ey: Y2, ez: Z2 } = apoint(other);
    const X1Z2 = M(X1 * Z2);
    const X2Z1 = M(X2 * Z1);
    const Y1Z2 = M(Y1 * Z2);
    const Y2Z1 = M(Y2 * Z1);
    return X1Z2 === X2Z1 && Y1Z2 === Y2Z1;
  }
  is0() {
    return this.equals(I);
  }
  negate() {
    return new Point(M(-this.ex), this.ey, this.ez, M(-this.et));
  }
  double() {
    const { ex: X1, ey: Y1, ez: Z1 } = this;
    const a = _a;
    const A = M(X1 * X1);
    const B = M(Y1 * Y1);
    const C2 = M(2n * M(Z1 * Z1));
    const D = M(a * A);
    const x1y1 = X1 + Y1;
    const E = M(M(x1y1 * x1y1) - A - B);
    const G = D + B;
    const F = G - C2;
    const H = D - B;
    const X3 = M(E * F);
    const Y3 = M(G * H);
    const T3 = M(E * H);
    const Z3 = M(F * G);
    return new Point(X3, Y3, Z3, T3);
  }
  add(other) {
    const { ex: X1, ey: Y1, ez: Z1, et: T1 } = this;
    const { ex: X2, ey: Y2, ez: Z2, et: T2 } = apoint(other);
    const a = _a;
    const d = _d;
    const A = M(X1 * X2);
    const B = M(Y1 * Y2);
    const C2 = M(T1 * d * T2);
    const D = M(Z1 * Z2);
    const E = M((X1 + Y1) * (X2 + Y2) - A - B);
    const F = M(D - C2);
    const G = M(D + C2);
    const H = M(B - a * A);
    const X3 = M(E * F);
    const Y3 = M(G * H);
    const T3 = M(E * H);
    const Z3 = M(F * G);
    return new Point(X3, Y3, Z3, T3);
  }
  multiply(n, safe = true) {
    if (!safe && (n === 0n || this.is0()))
      return I;
    arange(n, 1n, N);
    if (n === 1n)
      return this;
    if (this.equals(G))
      return wNAF(n).p;
    let p = I;
    let f = G;
    for (let d = this;n > 0n; d = d.double(), n >>= 1n) {
      if (n & 1n)
        p = p.add(d);
      else if (safe)
        f = f.add(d);
    }
    return p;
  }
  toAffine() {
    const { ex: x, ey: y, ez: z } = this;
    if (this.equals(I))
      return { x: 0n, y: 1n };
    const iz = invert(z, P);
    if (M(z * iz) !== 1n)
      err("invalid inverse");
    return { x: M(x * iz), y: M(y * iz) };
  }
  toBytes() {
    const { x, y } = this.assertValidity().toAffine();
    const b = numTo32bLE(y);
    b[31] |= x & 1n ? 128 : 0;
    return b;
  }
  toHex() {
    return bytesToHex(this.toBytes());
  }
  clearCofactor() {
    return this.multiply(big(h), false);
  }
  isSmallOrder() {
    return this.clearCofactor().is0();
  }
  isTorsionFree() {
    let p = this.multiply(N / 2n, false).double();
    if (N % 2n)
      p = p.add(this);
    return p.is0();
  }
  static fromHex(hex, zip215) {
    return Point.fromBytes(toU8(hex), zip215);
  }
  get x() {
    return this.toAffine().x;
  }
  get y() {
    return this.toAffine().y;
  }
  toRawBytes() {
    return this.toBytes();
  }
}
var G = new Point(Gx, Gy, 1n, M(Gx * Gy));
var I = new Point(0n, 1n, 1n, 0n);
Point.BASE = G;
Point.ZERO = I;
var numTo32bLE = (num) => hexToBytes(padh(arange(num, 0n, B256), L2)).reverse();
var bytesToNumLE = (b) => big("0x" + bytesToHex(u8fr(abytes(b)).reverse()));
var pow2 = (x, power) => {
  let r = x;
  while (power-- > 0n) {
    r *= r;
    r %= P;
  }
  return r;
};
var pow_2_252_3 = (x) => {
  const x2 = x * x % P;
  const b2 = x2 * x % P;
  const b4 = pow2(b2, 2n) * b2 % P;
  const b5 = pow2(b4, 1n) * x % P;
  const b10 = pow2(b5, 5n) * b5 % P;
  const b20 = pow2(b10, 10n) * b10 % P;
  const b40 = pow2(b20, 20n) * b20 % P;
  const b80 = pow2(b40, 40n) * b40 % P;
  const b160 = pow2(b80, 80n) * b80 % P;
  const b240 = pow2(b160, 80n) * b80 % P;
  const b250 = pow2(b240, 10n) * b10 % P;
  const pow_p_5_8 = pow2(b250, 2n) * x % P;
  return { pow_p_5_8, b2 };
};
var RM1 = 0x2b8324804fc1df0b2b4d00993dfbd7a72f431806ad2fe478c4ee1b274a0ea0b0n;
var uvRatio = (u, v) => {
  const v3 = M(v * v * v);
  const v7 = M(v3 * v3 * v);
  const pow = pow_2_252_3(u * v7).pow_p_5_8;
  let x = M(u * v3 * pow);
  const vx2 = M(v * x * x);
  const root1 = x;
  const root2 = M(x * RM1);
  const useRoot1 = vx2 === u;
  const useRoot2 = vx2 === M(-u);
  const noRoot = vx2 === M(-u * RM1);
  if (useRoot1)
    x = root1;
  if (useRoot2 || noRoot)
    x = root2;
  if ((M(x) & 1n) === 1n)
    x = M(-x);
  return { isValid: useRoot1 || useRoot2, value: x };
};
var modL_LE = (hash) => modN(bytesToNumLE(hash));
var sha512s = (...m) => callHash("sha512Sync")(...m);
var hash2extK = (hashed) => {
  const head = hashed.slice(0, L);
  head[0] &= 248;
  head[31] &= 127;
  head[31] |= 64;
  const prefix = hashed.slice(L, L2);
  const scalar = modL_LE(head);
  const point = G.multiply(scalar);
  const pointBytes = point.toBytes();
  return { head, prefix, scalar, point, pointBytes };
};
var getExtendedPublicKey = (priv) => hash2extK(sha512s(toU8(priv, L)));
var hashFinishS = (res) => res.finish(sha512s(res.hashable));
var _sign = (e, rBytes, msg) => {
  const { pointBytes: P2, scalar: s } = e;
  const r = modL_LE(rBytes);
  const R = G.multiply(r).toBytes();
  const hashable = concatBytes(R, P2, msg);
  const finish = (hashed) => {
    const S = modN(r + modL_LE(hashed) * s);
    return abytes(concatBytes(R, numTo32bLE(S)), L2);
  };
  return { hashable, finish };
};
var sign = (msg, privKey) => {
  const m = toU8(msg);
  const e = getExtendedPublicKey(privKey);
  const rBytes = sha512s(e.prefix, m);
  return hashFinishS(_sign(e, rBytes, m));
};
var etc = {
  sha512Async: async (...messages) => {
    const s = subtle();
    const m = concatBytes(...messages);
    return u8n(await s.digest("SHA-512", m.buffer));
  },
  sha512Sync: undefined,
  bytesToHex,
  hexToBytes,
  concatBytes,
  mod: M,
  invert,
  randomBytes
};
var W = 8;
var scalarBits = 256;
var pwindows = Math.ceil(scalarBits / W) + 1;
var pwindowSize = 2 ** (W - 1);
var precompute = () => {
  const points = [];
  let p = G;
  let b = p;
  for (let w = 0;w < pwindows; w++) {
    b = p;
    points.push(b);
    for (let i = 1;i < pwindowSize; i++) {
      b = b.add(p);
      points.push(b);
    }
    p = b.double();
  }
  return points;
};
var Gpows = undefined;
var ctneg = (cnd, p) => {
  const n = p.negate();
  return cnd ? n : p;
};
var wNAF = (n) => {
  const comp = Gpows || (Gpows = precompute());
  let p = I;
  let f = G;
  const pow_2_w = 2 ** W;
  const maxNum = pow_2_w;
  const mask = big(pow_2_w - 1);
  const shiftBy = big(W);
  for (let w = 0;w < pwindows; w++) {
    let wbits = Number(n & mask);
    n >>= shiftBy;
    if (wbits > pwindowSize) {
      wbits -= maxNum;
      n += 1n;
    }
    const off = w * pwindowSize;
    const offF = off;
    const offP = off + Math.abs(wbits) - 1;
    const isEven = w % 2 !== 0;
    const isNeg = wbits < 0;
    if (wbits === 0) {
      f = f.add(ctneg(isEven, comp[offF]));
    } else {
      p = p.add(ctneg(isNeg, comp[offP]));
    }
  }
  return { p, f };
};

// src/index.ts
import { createHash } from "crypto";
etc.sha512Sync = (...messages) => {
  const hash = createHash("sha512");
  for (const msg of messages) {
    hash.update(msg);
  }
  return new Uint8Array(hash.digest());
};

class MdsClient {
  serverUrl;
  clientId;
  privateKey;
  constructor(serverUrl, clientId, privateKey) {
    this.serverUrl = serverUrl.replace(/\/$/, "");
    this.clientId = clientId;
    this.privateKey = this.parsePrivateKey(privateKey);
  }
  parsePrivateKey(key) {
    if (key instanceof Uint8Array) {
      return this.normalizeKeyBytes(key);
    }
    try {
      const decoded = Buffer.from(key, "base64");
      const decodedStr = decoded.toString();
      if (decodedStr.includes("BEGIN OPENSSH PRIVATE KEY")) {
        return this.parseOpenSSHKey(decodedStr);
      }
      if (decoded.length === 32 || decoded.length === 64) {
        return this.normalizeKeyBytes(decoded);
      }
    } catch {}
    if (key.includes("BEGIN OPENSSH PRIVATE KEY")) {
      return this.parseOpenSSHKey(key);
    }
    if (/^[0-9a-fA-F]+$/.test(key) && (key.length === 64 || key.length === 128)) {
      const decoded = Buffer.from(key, "hex");
      return this.normalizeKeyBytes(decoded);
    }
    throw new Error("Invalid private key format. Expected Base64, hex, OpenSSH, or raw bytes (32 or 64 bytes)");
  }
  normalizeKeyBytes(bytes) {
    const arr = new Uint8Array(bytes);
    if (arr.length === 64) {
      return arr.slice(0, 32);
    }
    if (arr.length === 32) {
      return arr;
    }
    throw new Error(`Invalid key size: expected 32 or 64 bytes, got ${arr.length}`);
  }
  parseOpenSSHKey(pemContent) {
    const lines = pemContent.split(`
`);
    const bodyLines = lines.filter((line) => !line.startsWith("-----") && line.trim().length > 0);
    const body = Buffer.from(bodyLines.join(""), "base64");
    const magicPrefix = Buffer.from("openssh-key-v1\x00");
    if (!body.subarray(0, magicPrefix.length).equals(magicPrefix)) {
      throw new Error("Invalid OpenSSH key format");
    }
    const keyTypeStr = "ssh-ed25519";
    let idx = body.indexOf(keyTypeStr, magicPrefix.length);
    if (idx === -1) {
      throw new Error("Could not find Ed25519 key type in OpenSSH key");
    }
    idx = body.indexOf(keyTypeStr, idx + keyTypeStr.length);
    if (idx === -1) {
      throw new Error("Could not find private key section in OpenSSH key");
    }
    idx += keyTypeStr.length;
    idx += 4 + 32;
    const privkeyLen = body.readUInt32BE(idx);
    idx += 4;
    if (privkeyLen !== 64) {
      throw new Error(`Unexpected private key length: ${privkeyLen}`);
    }
    const fullPrivKey = body.subarray(idx, idx + 64);
    return new Uint8Array(fullPrivKey.subarray(0, 32));
  }
  async signRequest(method, path, body) {
    const timestamp = new Date().toISOString();
    const encoder = new TextEncoder;
    const bodyBytes = encoder.encode(body);
    const hashBuffer = await crypto.subtle.digest("SHA-256", bodyBytes);
    const bodyHash = Buffer.from(hashBuffer).toString("hex");
    const canonical = `${method}
${path}
${timestamp}
${bodyHash}`;
    const canonicalBytes = encoder.encode(canonical);
    const signature = sign(canonicalBytes, this.privateKey);
    const signatureB64 = Buffer.from(signature).toString("base64");
    return { timestamp, signature: signatureB64 };
  }
  async request(method, path, body) {
    const bodyStr = body ? JSON.stringify(body) : "";
    const { timestamp, signature } = await this.signRequest(method, path, bodyStr);
    const response = await fetch(`${this.serverUrl}${path}`, {
      method,
      headers: {
        "Content-Type": "application/json",
        "X-Client-Id": this.clientId,
        "X-Timestamp": timestamp,
        Authorization: `Signature ${signature}`
      },
      body: bodyStr || undefined
    });
    const responseBody = await response.text();
    let parsed;
    try {
      parsed = JSON.parse(responseBody);
    } catch {
      if (response.status === 202) {
        return { success: true, message: "Accepted" };
      }
      throw new Error(`API error: ${response.status} ${response.statusText}`);
    }
    if (response.status === 202 || response.status === 200) {
      return parsed;
    }
    const errorResp = parsed;
    if (errorResp?.error?.code) {
      throw new Error(`API error (${errorResp.error.code}): ${errorResp.error.message}`);
    }
    throw new Error(`API error: ${response.status} ${response.statusText}`);
  }
  async sendEmail(request) {
    return this.request("POST", "/v3/email", request);
  }
  async sendSms(request) {
    return this.request("POST", "/v3/sms", request);
  }
  async health() {
    const response = await fetch(`${this.serverUrl}/health`);
    if (!response.ok) {
      throw new Error(`Health check failed: ${response.status} ${response.statusText}`);
    }
    const text = await response.text();
    try {
      return JSON.parse(text);
    } catch {
      throw new Error(`Failed to parse JSON response: "${text.substring(0, 100)}..."`);
    }
  }
}
export {
  MdsClient
};
