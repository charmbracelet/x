#!/usr/bin/env lua
-- Extract LSP configurations from nvim-lspconfig and output JSON
-- Usage: lua extract-lsp-configs.lua <lsp-dir>

local lsp_dir = arg[1]
if not lsp_dir then
    io.stderr:write("Usage: lua extract-lsp-configs.lua <lsp-dir>\n")
    os.exit(1)
end

-- Mock lspconfig.util module
local lspconfig_util = {
    root_pattern = function(...)
        local patterns = {...}
        -- Return a function that returns nil (we capture patterns separately)
        return function() return nil end
    end,
    find_git_ancestor = function() return nil end,
    find_node_modules_ancestor = function() return nil end,
    find_package_json_ancestor = function() return nil end,
    insert_package_json = function() end,
    path = {
        exists = function() return false end,
        join = function(...) return table.concat({...}, "/") end,
        is_dir = function() return false end,
        is_file = function() return false end,
        dirname = function(p) return p:match("(.*/)[^/]*$") or "" end,
        is_absolute = function() return true end,
    },
    get_active_clients_list_by_ft = function() return {} end,
    get_active_client_by_name = function() return nil end,
}

-- Mock package.preload for lspconfig modules
package.preload['lspconfig.util'] = function() return lspconfig_util end
package.preload['lspconfig/util'] = function() return lspconfig_util end
package.preload['lspconfig.async'] = function() return { run = function() end } end
package.preload['lspconfig'] = function() return { util = lspconfig_util } end

local vim_fn = {
    system = function() return "" end,
    expand = function(s) return s end,
    fnamemodify = function(path, mods) return path end,
    filereadable = function() return 0 end,
    isdirectory = function() return 0 end,
    getcwd = function() return "/" end,
    executable = function() return 1 end,
    exepath = function(cmd) return cmd end,
    json_decode = function() return {} end,
    json_encode = function() return "{}" end,
    mkdir = function() return 0 end,
    glob = function() return "" end,
    globpath = function() return "" end,
    getenv = function() return "" end,
    has = function() return 0 end,
    stdpath = function() return "/tmp" end,
    shellescape = function(s) return "'" .. s .. "'" end,
    tempname = function() return "/tmp/nvim_temp" end,
    find = function() return {} end,  -- vim.fn.find returns a list
}

-- Mock vim APIs that configs might use
local vim_mock = {
    fn = setmetatable(vim_fn, {
        __index = function() return function() return "" end end
    }),
    api = setmetatable({}, {
        __index = function() return function() return {} end end
    }),
    env = setmetatable({}, {
        __index = function() return "" end
    }),
    fs = {
        root = function() return nil end,
        dirname = function(path) 
            if type(path) ~= "string" then return "" end
            return path:match("(.*/)[^/]*$") or "" 
        end,
        joinpath = function(...) return table.concat({...}, "/") end,
        normalize = function(p) return p or "" end,
        find = function() return {} end,
    },
    uv = {
        os_homedir = function() return os.getenv("HOME") or "/home/user" end,
        os_tmpdir = function() return os.getenv("TMPDIR") or "/tmp" end,
        fs_stat = function() return nil end,
        cwd = function() return "/" end,
    },
    lsp = {
        get_clients = function() return {} end,
        Config = {},  -- type annotation, not used at runtime
        rpc = {
            connect = function(host, port)
                -- Return a marker table so we know it's a TCP connection
                return { __tcp_connect = true, host = host, port = port }
            end,
        },
        protocol = {
            make_client_capabilities = function() return {} end,
            Methods = {},
            MessageType = { Error = 1, Warning = 2, Info = 3, Log = 4 },
        },
        config = setmetatable({}, {
            __index = function() return {} end,
            __call = function() return {} end,
        }),
    },
    g = setmetatable({}, {
        __index = function() return nil end,
        __newindex = function() end,
    }),
    loop = setmetatable({}, {
        __index = function() return function() return nil end end
    }),
    tbl_deep_extend = function(behavior, ...)
        local result = {}
        for _, t in ipairs({...}) do
            if type(t) == "table" then
                for k, v in pairs(t) do
                    result[k] = v
                end
            end
        end
        return result
    end,
    tbl_extend = function(behavior, ...)
        local result = {}
        for _, t in ipairs({...}) do
            if type(t) == "table" then
                for k, v in pairs(t) do
                    result[k] = v
                end
            end
        end
        return result
    end,
    notify = function() end,
    schedule = function(fn) end,
    system = function() end,
    trim = function(s) 
        if type(s) ~= "string" then return "" end
        return s:match("^%s*(.-)%s*$") 
    end,
    inspect = function(t) return tostring(t) end,
    startswith = function(s, prefix)
        if type(s) ~= "string" or type(prefix) ~= "string" then return false end
        return s:sub(1, #prefix) == prefix
    end,
    split = function(s, sep)
        local result = {}
        for match in (s..sep):gmatch("(.-)"..sep) do
            table.insert(result, match)
        end
        return result
    end,
    list_extend = function(dst, src)
        for _, v in ipairs(src) do
            table.insert(dst, v)
        end
        return dst
    end,
    deepcopy = function(t)
        if type(t) ~= "table" then return t end
        local copy = {}
        for k, v in pairs(t) do
            copy[k] = vim_mock.deepcopy(v)
        end
        return copy
    end,
    version = function() return { major = 0, minor = 11, patch = 0 } end,
    diagnostic = {
        severity = { ERROR = 1, WARN = 2, INFO = 3, HINT = 4 },
    },
    log = {
        levels = { DEBUG = 1, INFO = 2, WARN = 3, ERROR = 4 },
    },
    uri_from_bufnr = function() return "" end,
    uri_from_fname = function(fname) return "file://" .. fname end,
    empty_dict = function() return {} end,
    deprecate = function() end,
    find = function() return nil end,
    is_callable = function() return false end,
    o = setmetatable({}, {
        __index = function() return "" end,
    }),
    bo = setmetatable({}, {
        __index = function() return "" end,
    }),
    cmd = function() end,
    filetype = {
        match = function() return nil end,
    },
}
_G.vim = vim_mock

-- JSON encoding
local function json_encode_value(val)
    local t = type(val)
    if t == "nil" then
        return "null"
    elseif t == "boolean" then
        return val and "true" or "false"
    elseif t == "number" then
        return tostring(val)
    elseif t == "string" then
        -- Escape special characters
        local escaped = val:gsub('\\', '\\\\')
                           :gsub('"', '\\"')
                           :gsub('\n', '\\n')
                           :gsub('\r', '\\r')
                           :gsub('\t', '\\t')
        return '"' .. escaped .. '"'
    elseif t == "table" then
        -- Check if array (sequential integer keys starting at 1)
        local is_array = true
        local max_idx = 0
        for k, _ in pairs(val) do
            if type(k) ~= "number" or k ~= math.floor(k) or k < 1 then
                is_array = false
                break
            end
            if k > max_idx then max_idx = k end
        end
        if is_array and max_idx > 0 then
            -- Verify no holes
            for i = 1, max_idx do
                if val[i] == nil then
                    is_array = false
                    break
                end
            end
        end
        if max_idx == 0 then
            -- Empty table - check if it's meant to be an object
            is_array = false
        end
        
        if is_array then
            local parts = {}
            for i = 1, max_idx do
                parts[i] = json_encode_value(val[i])
            end
            return "[" .. table.concat(parts, ", ") .. "]"
        else
            local parts = {}
            for k, v in pairs(val) do
                if type(k) == "string" then
                    table.insert(parts, json_encode_value(k) .. ": " .. json_encode_value(v))
                end
            end
            table.sort(parts)  -- Consistent ordering
            return "{" .. table.concat(parts, ", ") .. "}"
        end
    elseif t == "function" then
        return nil  -- Skip functions
    else
        return nil
    end
end

-- Check if a value is "simple" (can be serialized to JSON without functions)
local function is_simple_value(val, depth)
    depth = depth or 0
    if depth > 10 then return false end
    
    local t = type(val)
    if t == "nil" or t == "boolean" or t == "number" or t == "string" then
        return true
    elseif t == "table" then
        for k, v in pairs(val) do
            if type(k) ~= "string" and type(k) ~= "number" then
                return false
            end
            if not is_simple_value(v, depth + 1) then
                return false
            end
        end
        return true
    end
    return false
end

-- Load and parse a config file
local function load_config(filepath)
    local chunk, err = loadfile(filepath)
    if not chunk then
        io.stderr:write("Error loading " .. filepath .. ": " .. tostring(err) .. "\n")
        return nil
    end
    
    local ok, result = pcall(chunk)
    if not ok then
        io.stderr:write("Error executing " .. filepath .. ": " .. tostring(result) .. "\n")
        return nil
    end
    
    if type(result) ~= "table" then
        return nil
    end
    
    return result
end

-- Extract relevant fields from config
local function extract_config(config)
    local result = {}
    
    -- Command (required)
    if config.cmd then
        if type(config.cmd) == "table" and #config.cmd > 0 then
            -- First element is the command
            result.command = config.cmd[1]
            -- Rest are args
            if #config.cmd > 1 then
                result.args = {}
                for i = 2, #config.cmd do
                    table.insert(result.args, config.cmd[i])
                end
            end
        elseif type(config.cmd) == "string" then
            result.command = config.cmd
        end
    end
    
    if not result.command then
        return nil
    end
    
    -- Filetypes
    if config.filetypes and type(config.filetypes) == "table" then
        local ft = {}
        for _, v in ipairs(config.filetypes) do
            if type(v) == "string" then
                table.insert(ft, v)
            end
        end
        if #ft > 0 then
            result.filetypes = ft
        end
    end
    
    -- Root markers
    if config.root_markers and type(config.root_markers) == "table" then
        local rm = {}
        for _, v in ipairs(config.root_markers) do
            if type(v) == "string" then
                table.insert(rm, v)
            end
        end
        if #rm > 0 then
            result.root_markers = rm
        end
    end
    
    -- Single file support
    if config.single_file_support ~= nil then
        result.single_file_support = config.single_file_support
    end
    
    -- Settings (if simple enough to serialize)
    if config.settings and type(config.settings) == "table" then
        if is_simple_value(config.settings) then
            -- Only include if not empty
            local has_content = false
            for _ in pairs(config.settings) do
                has_content = true
                break
            end
            if has_content then
                result.settings = config.settings
            end
        end
    end
    
    -- Init options
    if config.init_options and type(config.init_options) == "table" then
        if is_simple_value(config.init_options) then
            local has_content = false
            for _ in pairs(config.init_options) do
                has_content = true
                break
            end
            if has_content then
                result.init_options = config.init_options
            end
        end
    end
    
    return result
end

-- Get list of lua files
local function get_lua_files(dir)
    local files = {}
    local handle = io.popen('ls "' .. dir .. '"/*.lua 2>/dev/null')
    if handle then
        for line in handle:lines() do
            table.insert(files, line)
        end
        handle:close()
    end
    return files
end

-- Main
local files = get_lua_files(lsp_dir)
local configs = {}
local names = {}

for _, filepath in ipairs(files) do
    local name = filepath:match("([^/]+)%.lua$")
    if name then
        local config = load_config(filepath)
        if config then
            -- nvim-lspconfig wraps the actual config in default_config
            local actual_config = config.default_config or config
            local extracted = extract_config(actual_config)
            if extracted then
                configs[name] = extracted
                table.insert(names, name)
            end
        end
    end
end

-- Sort names for consistent output
table.sort(names)

-- Output JSON
io.write("{\n")
for i, name in ipairs(names) do
    local config = configs[name]
    local json = json_encode_value(config)
    if json then
        io.write('  "' .. name .. '": ' .. json)
        if i < #names then
            io.write(",")
        end
        io.write("\n")
    end
end
io.write("}\n")

io.stderr:write("Generated " .. #names .. " server configs\n")
