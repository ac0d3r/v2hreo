//
//  PopoverViewController.swift
//  v2hreo
//
//  Created by whoami on 2021/7/6.
//

import Cocoa

class PopoverViewController: NSViewController {
    @IBOutlet weak var v2appConfState: NSSwitch!
    
    @IBOutlet weak var v2appConfAddr: NSTextField!
    @IBOutlet weak var v2appConfPort: NSTextField!
    @IBOutlet weak var v2appConfProto: NSTextField!
    @IBOutlet weak var v2appConfSubAddr: NSTextField!
    
    @IBOutlet weak var v2appServerSelect: NSComboBox!
    
    var conf: Conf = Conf(addr:"", port:0, proto:"", subaddr:"")
    var links: [String] = []
    var select: Int = -1
    
    override func viewDidLoad() {
        super.viewDidLoad()
        v2appConfAddr.isEditable = false
        v2appConfPort.isEditable = false
        v2appConfProto.isEditable = false
        v2appConfSubAddr.placeholderString = "输入订阅地址"
        
        v2appConfState.state = NSControl.StateValue.off
        if appRunning(){
            v2appConfState.state = NSControl.StateValue.on
        }
        
        conf = appConfInfo()
        v2appConfAddr.stringValue = conf.addr
        v2appConfPort.stringValue = String(conf.port)
        v2appConfProto.stringValue = conf.proto
        v2appConfSubAddr.stringValue = conf.subaddr
        
        if conf.subaddr != ""{
            let err = appLoadSubAddr()
            if err != "" {
                print(err)
            }else{
                renderLinks()
            }
        }
    }

    @IBAction func v2appSwitch(_ sender: Any) {
        if v2appConfState.state == NSControl.StateValue.on && !appRunning(){
            printAppErr(err: appStart())
        }
        if v2appConfState.state == NSControl.StateValue.off && appRunning(){
            printAppErr(err: appClose())
        }
    }
    
    @IBAction func saveConf(_ sender: Any) {
        if conf.subaddr != v2appConfSubAddr.stringValue{
            conf.subaddr = v2appConfSubAddr.stringValue
            if !printAppErrWithCondition(err: appSetSubAddr(addr: conf.subaddr)) && !printAppErrWithCondition(err: appLoadSubAddr()){
                renderLinks()
            }
        }
    }

    @IBAction func selectServer(_ sender: Any) {
        if v2appServerSelect.indexOfSelectedItem >= 0 && links.count > 0 {
            if v2appServerSelect.indexOfSelectedItem != select{
                if !printAppErrWithCondition(err: appSelectLink(index: v2appServerSelect.indexOfSelectedItem)){
                    select = v2appServerSelect.indexOfSelectedItem
                }
            }
        }else{
            v2appServerSelect.selectItem(at:-1)
        }
    }

    @IBAction func v2appLoad(_ sender: Any) {
        if !printAppErrWithCondition(err: appLoadSubAddr()){
            renderLinks()
        }
    }
    
    @IBAction func pingServer(_ sender: Any) {
        if conf.subaddr != ""{
            if links.count <= 0 {
                printAppErr(err: appLoadSubAddr())
            }
            printAppErr(err: appPing())
            renderLinks()
        }
    }

    @IBAction func qutiBtn(_ sender: Any) {
        NSApplication.shared.terminate(self)
    }
    
    func renderLinks(){
        links = appListHosts()
        v2appServerSelect.removeAllItems()
        v2appServerSelect.addItems(withObjectValues: links)
    }
}

extension PopoverViewController {
    static func freshController() -> PopoverViewController {
        // 获取对Main.storyboard的引用
        let storyboard = NSStoryboard(name: NSStoryboard.Name("Main"), bundle: nil)
        // 为PopoverViewController创建一个标识符
        let identifier = NSStoryboard.SceneIdentifier("PopoverViewController")
        // 实例化PopoverViewController并返回
        guard let viewcontroller = storyboard.instantiateController(withIdentifier: identifier) as? PopoverViewController else {
            fatalError("Something Wrong with Main.storyboard")
        }
        return viewcontroller
    }
}

