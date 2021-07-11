//
//  V2rayssProxy.swift
//  v2hreo
//
//  Created by whoami on 2021/7/9.
//

import Foundation
import V2rayss


struct Conf: Codable{
    let addr: String
    let port: Int
    let proto: String
    var subaddr: String
}

func newApp() -> String
{
    return String(cString: V2rayss.newApp())
}

func appRunning() -> Bool
{
    let v = V2rayss.appRunning()
    if v == 1{
        return true
    }
    return false
}

func appConfInfo() -> Conf
{
    let jsonString = String(cString: V2rayss.appConfInfo())
    if jsonString.contains("err: "){
        print(jsonString)
        return Conf(addr: "", port: 0, proto: "", subaddr: "")
    }
    let jsonData = jsonString.data(using: .utf8)!
    do {
        let conf = try JSONDecoder().decode(Conf.self, from: jsonData)
        return conf
    }catch{
        return Conf(addr: "", port: 0, proto: "", subaddr: "")
    }
}

func appLoadSubAddr() -> String{
    return String(cString: V2rayss.appLoadSubAddr())
}

func appSetSubAddr(addr: String) -> String{
    let addrChar = UnsafeMutablePointer<Int8>(mutating: addr.cString(using: .utf8))
    if addrChar != nil{
        return String(cString: V2rayss.appSetSubAddr(addrChar))
    }
    return "err: String -> char*"
}

func appPing() -> String{
    return String(cString: V2rayss.appPing())
}

func appListHosts() -> [String]{
    let jsonString = String(cString: V2rayss.appListHosts())
    if jsonString.contains("err: "){
        print(jsonString)
        return [String]()
    }
    let jsonData = jsonString.data(using: .utf8)!
    do {
        let links = try JSONDecoder().decode([String].self, from: jsonData)
        return links
    }catch{
        return [String]()
    }
}

func appSelectLink(index: Int) -> String{
    return String(cString: V2rayss.appSelectLink(Int32(index)))
}

func appStart() -> String{
    return String(cString: V2rayss.appStart())
}

func appClose() -> String{
    return String(cString: V2rayss.appClose())
}

func printAppErrWithCondition(err: String) -> Bool {
    if err != ""{
        print(err)
        return true
    }
    return false
}

func printAppErr(err: String) {
    if err != ""{
        print(err)
    }
}
