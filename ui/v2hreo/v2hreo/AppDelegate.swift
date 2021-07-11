//
//  AppDelegate.swift
//  v2hreo
//
//  Created by whoami on 2021/7/2.
//

import Cocoa
import SwiftUI

@main
class AppDelegate: NSObject, NSApplicationDelegate {
    // 事件监控
    var eventMonitor: EventMonitor?
    // 状态栏
    let statusItem = NSStatusBar.system.statusItem(withLength: NSStatusItem.squareLength)
    // 控制Popover状态
    let popover = NSPopover()
    // user 数据
    let defaults = UserDefaults.standard
    @objc func togglePopover(_ sender: AnyObject) {
        if popover.isShown {
            closePopover(sender)
        } else {
            showPopover(sender)
        }
    }
    // 显示Popover
    @objc func showPopover(_ sender: AnyObject) {
        if let button = statusItem.button {
            popover.show(relativeTo: button.bounds, of: button, preferredEdge: NSRectEdge.minY)
        }
        eventMonitor?.start()
    }
    // 隐藏Popover
    @objc func closePopover(_ sender: AnyObject) {
        popover.performClose(sender)
        eventMonitor?.stop()
    }
    
    func applicationDidFinishLaunching(_ aNotification: Notification) {
        eventMonitor = EventMonitor(mask: [.leftMouseDown]) { [weak self] event in
          if let strongSelf = self, strongSelf.popover.isShown {
            strongSelf.closePopover(event!)
          }
        }
        if let button = statusItem.button {
            button.image = NSImage(named: "v2hreo")
            button.action = #selector(togglePopover(_:))
        }
        popover.contentViewController = PopoverViewController.freshController()
        //init v2rayss app
        if !printAppErrWithCondition(err: newApp()){
            // set subaddr
            let subaddr = defaults.string(forKey: "subaddr")
            if subaddr != nil{
                printAppErr(err: appSetSubAddr(addr: subaddr!))
            }
        }
    }

    func applicationWillTerminate(_ aNotification: Notification) {
        //close v2rayss app
        if appRunning(){
            printAppErr(err: appClose())
        }
        // save v2rayss subaddr
        let conf = appConfInfo()
        if conf.subaddr != ""{
            defaults.set(conf.subaddr, forKey: "subaddr")
        }
    }
}


