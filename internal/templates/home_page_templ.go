// Code generated by templ - DO NOT EDIT.

// templ: version: v0.2.793
package templates

//lint:file-ignore SA4006 This context is only used if a nested component is present.

import "github.com/a-h/templ"
import templruntime "github.com/a-h/templ/runtime"

type HomeViewModel struct {
}

func HomePage(vm HomeViewModel) templ.Component {
	return templruntime.GeneratedTemplate(func(templ_7745c5c3_Input templruntime.GeneratedComponentInput) (templ_7745c5c3_Err error) {
		templ_7745c5c3_W, ctx := templ_7745c5c3_Input.Writer, templ_7745c5c3_Input.Context
		if templ_7745c5c3_CtxErr := ctx.Err(); templ_7745c5c3_CtxErr != nil {
			return templ_7745c5c3_CtxErr
		}
		templ_7745c5c3_Buffer, templ_7745c5c3_IsBuffer := templruntime.GetBuffer(templ_7745c5c3_W)
		if !templ_7745c5c3_IsBuffer {
			defer func() {
				templ_7745c5c3_BufErr := templruntime.ReleaseBuffer(templ_7745c5c3_Buffer)
				if templ_7745c5c3_Err == nil {
					templ_7745c5c3_Err = templ_7745c5c3_BufErr
				}
			}()
		}
		ctx = templ.InitializeContext(ctx)
		templ_7745c5c3_Var1 := templ.GetChildren(ctx)
		if templ_7745c5c3_Var1 == nil {
			templ_7745c5c3_Var1 = templ.NopComponent
		}
		ctx = templ.ClearChildren(ctx)
		templ_7745c5c3_Var2 := templruntime.GeneratedTemplate(func(templ_7745c5c3_Input templruntime.GeneratedComponentInput) (templ_7745c5c3_Err error) {
			templ_7745c5c3_W, ctx := templ_7745c5c3_Input.Writer, templ_7745c5c3_Input.Context
			templ_7745c5c3_Buffer, templ_7745c5c3_IsBuffer := templruntime.GetBuffer(templ_7745c5c3_W)
			if !templ_7745c5c3_IsBuffer {
				defer func() {
					templ_7745c5c3_BufErr := templruntime.ReleaseBuffer(templ_7745c5c3_Buffer)
					if templ_7745c5c3_Err == nil {
						templ_7745c5c3_Err = templ_7745c5c3_BufErr
					}
				}()
			}
			ctx = templ.InitializeContext(ctx)
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<div class=\"relative flex flex-col items-center justify-center min-h-screen gap-5\"><div class=\"absolute flex top-0 right-0 z-10\"><a href=\"/stats\" class=\"inline-block p-4\">")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			templ_7745c5c3_Err = StatsIcon().Render(ctx, templ_7745c5c3_Buffer)
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("</a><div class=\"js-drawer-open p-4 cursor-pointer\">")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			templ_7745c5c3_Err = ConfigIcon().Render(ctx, templ_7745c5c3_Buffer)
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("</div></div><h1 class=\"text-3xl\">Generate a server name</h1><div id=\"generate-name-container\" class=\"w-full flex justify-center\"><span class=\"text-gray-400\">The name will be here</span></div><button hx-post=\"/generate\" hx-target=\"#generate-name-container\" class=\"group inline-block rounded-full bg-gradient-to-r from-purple-600 to-cyan-400 p-[2px] hover:text-white focus:outline-none focus:ring active:text-opacity-75\" type=\"submit\"><span class=\"block rounded-full bg-white px-8 py-3 text-sm font-medium group-hover:bg-transparent\">Generate</span></button></div><div id=\"drawer\" class=\"fixed top-0 right-0 z-20 w-60 h-screen p-4 bg-white overflow-y-auto transition-transform translate-x-full shadow-xl\" tabindex=\"-1\" aria-hiden=\"true\"><div class=\"relative flex flex-col w-full h-full\"><button type=\"button\" class=\"js-drawer-close text-gray-400 bg-transparent hover:bg-slate-300 hover:text-gray-900 rounded-lg text-sm w-6 h-6 absolute top-0 end-0 flex items-center justify-center\"><svg class=\"w-3 h-3\" aria-hidden=\"true\" xmlns=\"http://www.w3.org/2000/svg\" fill=\"none\" viewBox=\"0 0 14 14\"><path stroke=\"currentColor\" stroke-linecap=\"round\" stroke-linejoin=\"round\" stroke-width=\"2\" d=\"m1 1 6 6m0 0 6 6M7 7l6-6M7 7l-6 6\"></path></svg> <span class=\"sr-only\">Close menu</span></button><div class=\"text-xl font-semibold\">Configuration</div><div class=\"mt-4\"><div class=\"relative mb-6\"><label for=\"default-range\" class=\"block mb-2 text-sm font-medium text-gray-800\">Max Length</label> <input id=\"default-range\" type=\"range\" value=\"14\" min=\"7\" max=\"19\" class=\"accent-orange-600 bg-gradient-to-r from-purple-500 to-cyan-300 w-full h-2 rounded-lg appearance-none cursor-pointer\"> <span class=\"text-sm text-gray-500 dark:text-gray-400 absolute start-0 -bottom-6\">7</span> <span class=\"text-sm text-gray-500 dark:text-gray-400 absolute start-1/2 -translate-x-1/2 rtl:translate-x-1/2 -bottom-6\">12</span> <span class=\"text-sm text-gray-500 dark:text-gray-400 absolute end-0 -bottom-6\">19</span></div></div></div></div>")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			return templ_7745c5c3_Err
		})
		templ_7745c5c3_Err = Layout().Render(templ.WithChildren(ctx, templ_7745c5c3_Var2), templ_7745c5c3_Buffer)
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		return templ_7745c5c3_Err
	})
}

var _ = templruntime.GeneratedTemplate
