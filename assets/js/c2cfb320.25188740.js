"use strict";(self.webpackChunkwebsite=self.webpackChunkwebsite||[]).push([[429],{3905:(e,t,n)=>{n.d(t,{Zo:()=>c,kt:()=>m});var a=n(7294);function i(e,t,n){return t in e?Object.defineProperty(e,t,{value:n,enumerable:!0,configurable:!0,writable:!0}):e[t]=n,e}function r(e,t){var n=Object.keys(e);if(Object.getOwnPropertySymbols){var a=Object.getOwnPropertySymbols(e);t&&(a=a.filter((function(t){return Object.getOwnPropertyDescriptor(e,t).enumerable}))),n.push.apply(n,a)}return n}function o(e){for(var t=1;t<arguments.length;t++){var n=null!=arguments[t]?arguments[t]:{};t%2?r(Object(n),!0).forEach((function(t){i(e,t,n[t])})):Object.getOwnPropertyDescriptors?Object.defineProperties(e,Object.getOwnPropertyDescriptors(n)):r(Object(n)).forEach((function(t){Object.defineProperty(e,t,Object.getOwnPropertyDescriptor(n,t))}))}return e}function s(e,t){if(null==e)return{};var n,a,i=function(e,t){if(null==e)return{};var n,a,i={},r=Object.keys(e);for(a=0;a<r.length;a++)n=r[a],t.indexOf(n)>=0||(i[n]=e[n]);return i}(e,t);if(Object.getOwnPropertySymbols){var r=Object.getOwnPropertySymbols(e);for(a=0;a<r.length;a++)n=r[a],t.indexOf(n)>=0||Object.prototype.propertyIsEnumerable.call(e,n)&&(i[n]=e[n])}return i}var l=a.createContext({}),d=function(e){var t=a.useContext(l),n=t;return e&&(n="function"==typeof e?e(t):o(o({},t),e)),n},c=function(e){var t=d(e.components);return a.createElement(l.Provider,{value:t},e.children)},p="mdxType",u={inlineCode:"code",wrapper:function(e){var t=e.children;return a.createElement(a.Fragment,{},t)}},h=a.forwardRef((function(e,t){var n=e.components,i=e.mdxType,r=e.originalType,l=e.parentName,c=s(e,["components","mdxType","originalType","parentName"]),p=d(n),h=i,m=p["".concat(l,".").concat(h)]||p[h]||u[h]||r;return n?a.createElement(m,o(o({ref:t},c),{},{components:n})):a.createElement(m,o({ref:t},c))}));function m(e,t){var n=arguments,i=t&&t.mdxType;if("string"==typeof e||i){var r=n.length,o=new Array(r);o[0]=h;var s={};for(var l in t)hasOwnProperty.call(t,l)&&(s[l]=t[l]);s.originalType=e,s[p]="string"==typeof e?e:i,o[1]=s;for(var d=2;d<r;d++)o[d]=n[d];return a.createElement.apply(null,o)}return a.createElement.apply(null,n)}h.displayName="MDXCreateElement"},6096:(e,t,n)=>{n.r(t),n.d(t,{assets:()=>l,contentTitle:()=>o,default:()=>u,frontMatter:()=>r,metadata:()=>s,toc:()=>d});var a=n(7462),i=(n(7294),n(3905));const r={sidebar_position:4},o="Validator instructions for Changeover Procedure",s={unversionedId:"validators/changeover-procedure",id:"validators/changeover-procedure",title:"Validator instructions for Changeover Procedure",description:"More details available in Changeover Procedure documentation.",source:"@site/docs/validators/changeover-procedure.md",sourceDirName:"validators",slug:"/validators/changeover-procedure",permalink:"/interchain-security/validators/changeover-procedure",draft:!1,tags:[],version:"current",sidebarPosition:4,frontMatter:{sidebar_position:4},sidebar:"tutorialSidebar",previous:{title:"Withdrawing consumer chain validator rewards",permalink:"/interchain-security/validators/withdraw_rewards"},next:{title:"Joining Neutron",permalink:"/interchain-security/validators/joining-neutron"}},l={},d=[{value:"Timeline",id:"timeline",level:2},{value:"1. <code>ConsumerAdditionProposal</code> on provider chain",id:"1-consumeradditionproposal-on-provider-chain",level:3},{value:"2. <code>SoftwareUpgradeProposal</code> on the standalone/consumer chain",id:"2-softwareupgradeproposal-on-the-standaloneconsumer-chain",level:3},{value:"3. Assigning a consumer key",id:"3-assigning-a-consumer-key",level:3},{value:"4. Perform the software ugprade on standalone chain",id:"4-perform-the-software-ugprade-on-standalone-chain",level:3},{value:"FAQ",id:"faq",level:2},{value:"Can I reuse the same validator key for the <code>consumer</code> chain that I am already using on the <code>standalone</code> chain? Will I need to perform a <code>AssignConsumerKey</code> tx with this key before spawn time?",id:"can-i-reuse-the-same-validator-key-for-the-consumer-chain-that-i-am-already-using-on-the-standalone-chain-will-i-need-to-perform-a-assignconsumerkey-tx-with-this-key-before-spawn-time",level:3},{value:"Can I continue using the same node that was validating the <code>standalone</code> chain?",id:"can-i-continue-using-the-same-node-that-was-validating-the-standalone-chain",level:3},{value:"Can I set up a new node to validate the <code>standalone/consumer</code> chain after it transitions to replicated security?",id:"can-i-set-up-a-new-node-to-validate-the-standaloneconsumer-chain-after-it-transitions-to-replicated-security",level:3},{value:"What happens to the <code>standalone</code> validator set after it after it transitions to replicated security?",id:"what-happens-to-the-standalone-validator-set-after-it-after-it-transitions-to-replicated-security",level:3},{value:"Credits",id:"credits",level:2}],c={toc:d},p="wrapper";function u(e){let{components:t,...r}=e;return(0,i.kt)(p,(0,a.Z)({},c,r,{components:t,mdxType:"MDXLayout"}),(0,i.kt)("h1",{id:"validator-instructions-for-changeover-procedure"},"Validator instructions for Changeover Procedure"),(0,i.kt)("p",null,"More details available in ",(0,i.kt)("a",{parentName:"p",href:"/interchain-security/consumer-development/changeover-procedure"},"Changeover Procedure documentation"),"."),(0,i.kt)("p",null,"A major difference betwen launching a new consumer chain vs. onboarding a standalone chain to ICS is that there is no consumer genesis available for the standalone chain. Since a standalone chain already exists, its state must be preserved once it transitions to being a consumer chain."),(0,i.kt)("h2",{id:"timeline"},"Timeline"),(0,i.kt)("p",null,"Upgrading standalone chains can be best visualised using a timeline, such as the one available ",(0,i.kt)("a",{parentName:"p",href:"https://app.excalidraw.com/l/9UFOCMAZLAI/5EVLj0WJcwt"},"Excalidraw graphic by Stride"),"."),(0,i.kt)("p",null,"There is some flexibility with regards to how the changeover procedure is executed, so please make sure to follow the guides provided by the team doing the changeover."),(0,i.kt)("p",null,(0,i.kt)("img",{alt:"Standalone to consumer transition timeline",src:n(6038).Z,width:"5307",height:"2157"})),(0,i.kt)("h3",{id:"1-consumeradditionproposal-on-provider-chain"},"1. ",(0,i.kt)("inlineCode",{parentName:"h3"},"ConsumerAdditionProposal")," on provider chain"),(0,i.kt)("p",null,"This step will add the standalone chain to the list of consumer chains secured by the provider.\nThis step dictates the ",(0,i.kt)("inlineCode",{parentName:"p"},"spawn_time"),". After ",(0,i.kt)("inlineCode",{parentName:"p"},"spawn_time")," the CCV state (initial validator set of the provider) will be available to the consumer."),(0,i.kt)("p",null,"To obtain it from the provider use:"),(0,i.kt)("pre",null,(0,i.kt)("code",{parentName:"pre",className:"language-bash"},"gaiad q provider consumer-genesis stride-1 -o json > ccv-state.json\njq -s '.[0].app_state.ccvconsumer = .[1] | .[0]' genesis.json ccv-state.json > ccv.json\n")),(0,i.kt)("h3",{id:"2-softwareupgradeproposal-on-the-standaloneconsumer-chain"},"2. ",(0,i.kt)("inlineCode",{parentName:"h3"},"SoftwareUpgradeProposal")," on the standalone/consumer chain"),(0,i.kt)("p",null,"This upgrade proposal will introduce ICS to the standalone chain, making it a consumer."),(0,i.kt)("h3",{id:"3-assigning-a-consumer-key"},"3. Assigning a consumer key"),(0,i.kt)("p",null,"After ",(0,i.kt)("inlineCode",{parentName:"p"},"spawn_time"),", make sure to assign a consumer key if you intend to use one."),(0,i.kt)("p",null,"Instructions are available ",(0,i.kt)("a",{parentName:"p",href:"/interchain-security/features/key-assignment"},"here")),(0,i.kt)("h3",{id:"4-perform-the-software-ugprade-on-standalone-chain"},"4. Perform the software ugprade on standalone chain"),(0,i.kt)("p",null,"Please use instructions provided by the standalone chain team and make sure to reach out if you are facing issues.\nThe upgrade preparation depends on your setup, so please make sure you prepare ahead of time."),(0,i.kt)("admonition",{type:"danger"},(0,i.kt)("p",{parentName:"admonition"},"The ",(0,i.kt)("inlineCode",{parentName:"p"},"ccv.json")," from step 1. must be made available on the machine running the standalone/consumer chain at standalone chain ",(0,i.kt)("inlineCode",{parentName:"p"},"upgrade_height"),". This file contains the initial validator set and parameters required for normal ICS operation."),(0,i.kt)("p",{parentName:"admonition"},"Usually, the file is placed in ",(0,i.kt)("inlineCode",{parentName:"p"},"$NODE_HOME/config")," but this is not a strict requirement. The exact details are available in the upgrade code of the standalone/consumer chain.")),(0,i.kt)("p",null,(0,i.kt)("strong",{parentName:"p"},"Performing this upgrade will transition the standalone chain to be a consumer chain.")),(0,i.kt)("p",null,'After 3 blocks, the standalone chain will stop using the "old" validator set and begin using the ',(0,i.kt)("inlineCode",{parentName:"p"},"provider")," validator set."),(0,i.kt)("h2",{id:"faq"},"FAQ"),(0,i.kt)("h3",{id:"can-i-reuse-the-same-validator-key-for-the-consumer-chain-that-i-am-already-using-on-the-standalone-chain-will-i-need-to-perform-a-assignconsumerkey-tx-with-this-key-before-spawn-time"},"Can I reuse the same validator key for the ",(0,i.kt)("inlineCode",{parentName:"h3"},"consumer")," chain that I am already using on the ",(0,i.kt)("inlineCode",{parentName:"h3"},"standalone")," chain? Will I need to perform a ",(0,i.kt)("inlineCode",{parentName:"h3"},"AssignConsumerKey")," tx with this key before spawn time?"),(0,i.kt)("p",null,"Validators must either assign a key or use the same key as on the ",(0,i.kt)("inlineCode",{parentName:"p"},"provider"),"."),(0,i.kt)("p",null,"If you are validating both the ",(0,i.kt)("inlineCode",{parentName:"p"},"standalone")," and the ",(0,i.kt)("inlineCode",{parentName:"p"},"provider"),", you ",(0,i.kt)("strong",{parentName:"p"},"can")," use your current ",(0,i.kt)("inlineCode",{parentName:"p"},"standalone")," key with some caveats:"),(0,i.kt)("ul",null,(0,i.kt)("li",{parentName:"ul"},"you must submit an ",(0,i.kt)("inlineCode",{parentName:"li"},"AssignConsumerKey")," tx with your current ",(0,i.kt)("inlineCode",{parentName:"li"},"standalone")," validator key"),(0,i.kt)("li",{parentName:"ul"},"it is best to submit ",(0,i.kt)("inlineCode",{parentName:"li"},"AssignConsumerKey")," tx before ",(0,i.kt)("inlineCode",{parentName:"li"},"spawn_time")),(0,i.kt)("li",{parentName:"ul"},"if you do not submit the Tx, it is assumed that you will be re-using your ",(0,i.kt)("inlineCode",{parentName:"li"},"provider")," key to validate the ",(0,i.kt)("inlineCode",{parentName:"li"},"standalone/consumer")," chain")),(0,i.kt)("h3",{id:"can-i-continue-using-the-same-node-that-was-validating-the-standalone-chain"},"Can I continue using the same node that was validating the ",(0,i.kt)("inlineCode",{parentName:"h3"},"standalone")," chain?"),(0,i.kt)("p",null,"Yes."),(0,i.kt)("p",null,"Please assign your consensus key as stated aboce."),(0,i.kt)("h3",{id:"can-i-set-up-a-new-node-to-validate-the-standaloneconsumer-chain-after-it-transitions-to-replicated-security"},"Can I set up a new node to validate the ",(0,i.kt)("inlineCode",{parentName:"h3"},"standalone/consumer")," chain after it transitions to replicated security?"),(0,i.kt)("p",null,"Yes."),(0,i.kt)("p",null,"If you are planning to do this please make sure that the node is synced with ",(0,i.kt)("inlineCode",{parentName:"p"},"standalone")," network and to submit ",(0,i.kt)("inlineCode",{parentName:"p"},"AssignConsumerKey")," tx before ",(0,i.kt)("inlineCode",{parentName:"p"},"spawn_time"),"."),(0,i.kt)("h3",{id:"what-happens-to-the-standalone-validator-set-after-it-after-it-transitions-to-replicated-security"},"What happens to the ",(0,i.kt)("inlineCode",{parentName:"h3"},"standalone")," validator set after it after it transitions to replicated security?"),(0,i.kt)("p",null,"The ",(0,i.kt)("inlineCode",{parentName:"p"},"standalone")," chain validators will stop being validators after the first 3 blocks are created while using replicated security. The ",(0,i.kt)("inlineCode",{parentName:"p"},"standalone")," validators will become ",(0,i.kt)("strong",{parentName:"p"},"governors")," and still can receive delegations if the ",(0,i.kt)("inlineCode",{parentName:"p"},"consumer")," chain is using the ",(0,i.kt)("inlineCode",{parentName:"p"},"consumer-democracy")," module."),(0,i.kt)("p",null,(0,i.kt)("strong",{parentName:"p"},"Governors DO NOT VALIDATE BLOCKS"),"."),(0,i.kt)("p",null,"Instead, they can participate in the governance process and take on other chain-specific roles."),(0,i.kt)("h2",{id:"credits"},"Credits"),(0,i.kt)("p",null,"Thank you Stride team for providing detailed instructions about the changeover procedure."))}u.isMDXComponent=!0},6038:(e,t,n)=>{n.d(t,{Z:()=>a});const a=n.p+"assets/images/ics_changeover_timeline_stride-9bcad1834fef24a0fea7f2c80c9ccd71.png"}}]);