export namespace dns {
	
	export class ServiceStatus {
	    installed: boolean;
	    running: boolean;
	    resolver: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ServiceStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.installed = source["installed"];
	        this.running = source["running"];
	        this.resolver = source["resolver"];
	    }
	}

}

export namespace log {
	
	export class LogFile {
	    name: string;
	    path: string;
	    size: number;
	
	    static createFrom(source: any = {}) {
	        return new LogFile(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.path = source["path"];
	        this.size = source["size"];
	    }
	}

}

export namespace main {
	
	export class DashboardStatus {
	    nginx_status: nginx.ServiceStatus;
	    dns_status: dns.ServiceStatus;
	    fpm_status: php.FpmStatus;
	    php_versions: number;
	    node_versions: number;
	    sites_count: number;
	    active_php: string;
	    active_node: string;
	    first_run: boolean;
	    debug_mode: boolean;
	    shell_integrated: boolean;
	
	    static createFrom(source: any = {}) {
	        return new DashboardStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.nginx_status = this.convertValues(source["nginx_status"], nginx.ServiceStatus);
	        this.dns_status = this.convertValues(source["dns_status"], dns.ServiceStatus);
	        this.fpm_status = this.convertValues(source["fpm_status"], php.FpmStatus);
	        this.php_versions = source["php_versions"];
	        this.node_versions = source["node_versions"];
	        this.sites_count = source["sites_count"];
	        this.active_php = source["active_php"];
	        this.active_node = source["active_node"];
	        this.first_run = source["first_run"];
	        this.debug_mode = source["debug_mode"];
	        this.shell_integrated = source["shell_integrated"];
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

export namespace nginx {
	
	export class ServiceStatus {
	    installed: boolean;
	    running: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ServiceStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.installed = source["installed"];
	        this.running = source["running"];
	    }
	}

}

export namespace node {
	
	export class NodeVersion {
	    version: string;
	    path: string;
	    active: boolean;
	
	    static createFrom(source: any = {}) {
	        return new NodeVersion(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.version = source["version"];
	        this.path = source["path"];
	        this.active = source["active"];
	    }
	}

}

export namespace php {
	
	export class AvailablePhpVersion {
	    version: string;
	    series: string;
	    internal_version: number;
	    installed: boolean;
	
	    static createFrom(source: any = {}) {
	        return new AvailablePhpVersion(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.version = source["version"];
	        this.series = source["series"];
	        this.internal_version = source["internal_version"];
	        this.installed = source["installed"];
	    }
	}
	export class FpmStatus {
	    running: boolean;
	    version: string;
	    socket: string;
	
	    static createFrom(source: any = {}) {
	        return new FpmStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.running = source["running"];
	        this.version = source["version"];
	        this.socket = source["socket"];
	    }
	}
	export class PhpVersion {
	    version: string;
	    path: string;
	    binary: string;
	    fpm_bin: string;
	    active: boolean;
	    source: string;
	
	    static createFrom(source: any = {}) {
	        return new PhpVersion(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.version = source["version"];
	        this.path = source["path"];
	        this.binary = source["binary"];
	        this.fpm_bin = source["fpm_bin"];
	        this.active = source["active"];
	        this.source = source["source"];
	    }
	}

}

export namespace site {
	
	export class Site {
	    domain: string;
	    path: string;
	    php_version: string;
	    ssl_enabled: boolean;
	    created_at: string;
	
	    static createFrom(source: any = {}) {
	        return new Site(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.domain = source["domain"];
	        this.path = source["path"];
	        this.php_version = source["php_version"];
	        this.ssl_enabled = source["ssl_enabled"];
	        this.created_at = source["created_at"];
	    }
	}

}

