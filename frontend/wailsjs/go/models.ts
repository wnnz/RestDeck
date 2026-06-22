export namespace domain {
	
	export class AuthConfig {
	    type: string;
	    values: Record<string, string>;
	
	    static createFrom(source: any = {}) {
	        return new AuthConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.type = source["type"];
	        this.values = source["values"];
	    }
	}
	export class ProxyConfig {
	    mode: string;
	    url: string;
	
	    static createFrom(source: any = {}) {
	        return new ProxyConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.mode = source["mode"];
	        this.url = source["url"];
	    }
	}
	export class FormItem {
	    id: string;
	    enabled: boolean;
	    key: string;
	    type: string;
	    value: string;
	    filePath: string;
	    description: string;
	
	    static createFrom(source: any = {}) {
	        return new FormItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.enabled = source["enabled"];
	        this.key = source["key"];
	        this.type = source["type"];
	        this.value = source["value"];
	        this.filePath = source["filePath"];
	        this.description = source["description"];
	    }
	}
	export class KeyValue {
	    id: string;
	    enabled: boolean;
	    key: string;
	    value: string;
	    description: string;
	    secret: boolean;
	    valueType: string;
	    timestampFormat: string;
	    sourceRequestId: string;
	    jsonPath: string;
	    responseStrategy: string;
	    refreshAfterSeconds: number;
	    fallbackValue: string;
	
	    static createFrom(source: any = {}) {
	        return new KeyValue(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.enabled = source["enabled"];
	        this.key = source["key"];
	        this.value = source["value"];
	        this.description = source["description"];
	        this.secret = source["secret"];
	        this.valueType = source["valueType"];
	        this.timestampFormat = source["timestampFormat"];
	        this.sourceRequestId = source["sourceRequestId"];
	        this.jsonPath = source["jsonPath"];
	        this.responseStrategy = source["responseStrategy"];
	        this.refreshAfterSeconds = source["refreshAfterSeconds"];
	        this.fallbackValue = source["fallbackValue"];
	    }
	}
	export class Request {
	    id: string;
	    collectionId: string;
	    parentId: string;
	    name: string;
	    method: string;
	    url: string;
	    params: KeyValue[];
	    headers: KeyValue[];
	    bodyMode: string;
	    body: string;
	    formItems: FormItem[];
	    auth: AuthConfig;
	    proxy: ProxyConfig;
	    preScript: string;
	    testScript: string;
	    timeoutMs: number;
	    sortOrder: number;
	    // Go type: time
	    updatedAt: any;
	
	    static createFrom(source: any = {}) {
	        return new Request(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.collectionId = source["collectionId"];
	        this.parentId = source["parentId"];
	        this.name = source["name"];
	        this.method = source["method"];
	        this.url = source["url"];
	        this.params = this.convertValues(source["params"], KeyValue);
	        this.headers = this.convertValues(source["headers"], KeyValue);
	        this.bodyMode = source["bodyMode"];
	        this.body = source["body"];
	        this.formItems = this.convertValues(source["formItems"], FormItem);
	        this.auth = this.convertValues(source["auth"], AuthConfig);
	        this.proxy = this.convertValues(source["proxy"], ProxyConfig);
	        this.preScript = source["preScript"];
	        this.testScript = source["testScript"];
	        this.timeoutMs = source["timeoutMs"];
	        this.sortOrder = source["sortOrder"];
	        this.updatedAt = this.convertValues(source["updatedAt"], null);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Folder {
	    id: string;
	    collectionId: string;
	    parentId: string;
	    name: string;
	    sortOrder: number;
	    // Go type: time
	    updatedAt: any;
	
	    static createFrom(source: any = {}) {
	        return new Folder(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.collectionId = source["collectionId"];
	        this.parentId = source["parentId"];
	        this.name = source["name"];
	        this.sortOrder = source["sortOrder"];
	        this.updatedAt = this.convertValues(source["updatedAt"], null);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Collection {
	    id: string;
	    name: string;
	    description: string;
	    folders: Folder[];
	    requests: Request[];
	    // Go type: time
	    createdAt: any;
	    // Go type: time
	    updatedAt: any;
	
	    static createFrom(source: any = {}) {
	        return new Collection(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.folders = this.convertValues(source["folders"], Folder);
	        this.requests = this.convertValues(source["requests"], Request);
	        this.createdAt = this.convertValues(source["createdAt"], null);
	        this.updatedAt = this.convertValues(source["updatedAt"], null);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Cookie {
	    name: string;
	    value: string;
	    domain: string;
	    path: string;
	    // Go type: time
	    expires: any;
	    httpOnly: boolean;
	    secure: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Cookie(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.value = source["value"];
	        this.domain = source["domain"];
	        this.path = source["path"];
	        this.expires = this.convertValues(source["expires"], null);
	        this.httpOnly = source["httpOnly"];
	        this.secure = source["secure"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Environment {
	    id: string;
	    name: string;
	    variables: KeyValue[];
	    isActive: boolean;
	    // Go type: time
	    updatedAt: any;
	
	    static createFrom(source: any = {}) {
	        return new Environment(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.variables = this.convertValues(source["variables"], KeyValue);
	        this.isActive = source["isActive"];
	        this.updatedAt = this.convertValues(source["updatedAt"], null);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	
	export class TestResult {
	    name: string;
	    passed: boolean;
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new TestResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.passed = source["passed"];
	        this.message = source["message"];
	    }
	}
	export class Response {
	    statusCode: number;
	    status: string;
	    durationMs: number;
	    sizeBytes: number;
	    headers: KeyValue[];
	    cookies: Cookie[];
	    body: string;
	    contentType: string;
	    testResults: TestResult[];
	    error: string;
	    requestedUrl: string;
	
	    static createFrom(source: any = {}) {
	        return new Response(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.statusCode = source["statusCode"];
	        this.status = source["status"];
	        this.durationMs = source["durationMs"];
	        this.sizeBytes = source["sizeBytes"];
	        this.headers = this.convertValues(source["headers"], KeyValue);
	        this.cookies = this.convertValues(source["cookies"], Cookie);
	        this.body = source["body"];
	        this.contentType = source["contentType"];
	        this.testResults = this.convertValues(source["testResults"], TestResult);
	        this.error = source["error"];
	        this.requestedUrl = source["requestedUrl"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class HistoryItem {
	    id: string;
	    requestId: string;
	    name: string;
	    method: string;
	    url: string;
	    statusCode: number;
	    durationMs: number;
	    // Go type: time
	    createdAt: any;
	    request: Request;
	    response: Response;
	
	    static createFrom(source: any = {}) {
	        return new HistoryItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.requestId = source["requestId"];
	        this.name = source["name"];
	        this.method = source["method"];
	        this.url = source["url"];
	        this.statusCode = source["statusCode"];
	        this.durationMs = source["durationMs"];
	        this.createdAt = this.convertValues(source["createdAt"], null);
	        this.request = this.convertValues(source["request"], Request);
	        this.response = this.convertValues(source["response"], Response);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	
	
	
	export class RunnerResult {
	    id: string;
	    collectionId: string;
	    environmentId: string;
	    name: string;
	    iterations: number;
	    passed: number;
	    failed: number;
	    durationMs: number;
	    items: TestResult[];
	    // Go type: time
	    createdAt: any;
	
	    static createFrom(source: any = {}) {
	        return new RunnerResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.collectionId = source["collectionId"];
	        this.environmentId = source["environmentId"];
	        this.name = source["name"];
	        this.iterations = source["iterations"];
	        this.passed = source["passed"];
	        this.failed = source["failed"];
	        this.durationMs = source["durationMs"];
	        this.items = this.convertValues(source["items"], TestResult);
	        this.createdAt = this.convertValues(source["createdAt"], null);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Settings {
	    language: string;
	    theme: string;
	    defaultProxy: ProxyConfig;
	
	    static createFrom(source: any = {}) {
	        return new Settings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.language = source["language"];
	        this.theme = source["theme"];
	        this.defaultProxy = this.convertValues(source["defaultProxy"], ProxyConfig);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	export class WorkspaceState {
	    collections: Collection[];
	    environments: Environment[];
	    history: HistoryItem[];
	    globals: KeyValue[];
	    activeEnvironmentId: string;
	    settings: Settings;
	
	    static createFrom(source: any = {}) {
	        return new WorkspaceState(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.collections = this.convertValues(source["collections"], Collection);
	        this.environments = this.convertValues(source["environments"], Environment);
	        this.history = this.convertValues(source["history"], HistoryItem);
	        this.globals = this.convertValues(source["globals"], KeyValue);
	        this.activeEnvironmentId = source["activeEnvironmentId"];
	        this.settings = this.convertValues(source["settings"], Settings);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

export namespace realtime {
	
	export class SSERequest {
	    url: string;
	    headers: domain.KeyValue[];
	    proxy: domain.ProxyConfig;
	    timeoutMs: number;
	    maxEvents: number;
	
	    static createFrom(source: any = {}) {
	        return new SSERequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.url = source["url"];
	        this.headers = this.convertValues(source["headers"], domain.KeyValue);
	        this.proxy = this.convertValues(source["proxy"], domain.ProxyConfig);
	        this.timeoutMs = source["timeoutMs"];
	        this.maxEvents = source["maxEvents"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class SSEResult {
	    statusCode: number;
	    events: string[];
	    durationMs: number;
	    error: string;
	
	    static createFrom(source: any = {}) {
	        return new SSEResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.statusCode = source["statusCode"];
	        this.events = source["events"];
	        this.durationMs = source["durationMs"];
	        this.error = source["error"];
	    }
	}
	export class WebSocketRequest {
	    url: string;
	    message: string;
	    headers: domain.KeyValue[];
	    proxy: domain.ProxyConfig;
	    timeoutMs: number;
	
	    static createFrom(source: any = {}) {
	        return new WebSocketRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.url = source["url"];
	        this.message = source["message"];
	        this.headers = this.convertValues(source["headers"], domain.KeyValue);
	        this.proxy = this.convertValues(source["proxy"], domain.ProxyConfig);
	        this.timeoutMs = source["timeoutMs"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class WebSocketResult {
	    connected: boolean;
	    sent: string;
	    received: string[];
	    durationMs: number;
	    error: string;
	
	    static createFrom(source: any = {}) {
	        return new WebSocketResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.connected = source["connected"];
	        this.sent = source["sent"];
	        this.received = source["received"];
	        this.durationMs = source["durationMs"];
	        this.error = source["error"];
	    }
	}

}

