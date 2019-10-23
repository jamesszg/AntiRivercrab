function FindProxyForURL(url, host)
{
    if (shExpMatch(host, "*.ppgame.com") && (shExpMatch(url, "*/Index/index") || shExpMatch(url, "*/Index/getDigitalSkyNbUid"))) 
        return "PROXY ar.xuanxuan.tech:8888";
    return "DIRECT";
}

