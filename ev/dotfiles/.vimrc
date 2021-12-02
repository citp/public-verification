"""TODO: make a shortcut to backspace all whitespace at beginning of line
""" and end up at end of prev line

"""TODO: Move neovim-only functionality to neovim config file
"""TODO: Move plugin-specific stuff to folders in .vim/plugged/

"""TODO: Keep hidden modified scratch buffer around when there are multiple buffers, to prevent exiting.

filetype plugin indent on
" show existing tab with 4 spaces width
set tabstop=4
" backspace 4 spaces at a time when tab
set softtabstop=4
" when indenting with '>', use 4 spaces width
set shiftwidth=4
" On pressing tab, insert 4 spaces
set expandtab
" Show matching brackets
set showmatch
" Do not wrap text that is too long
set nowrap
" Allow editing multiple buffers without saving
set hidden

" Redefine <leader> to ',' because '\' is really far away
let mapleader = ","

" Always generate a filename when using grep (even for a single file)
set grepprg=grep\ -nH\ $*

" Editing empty .tex files is done as tex instead of plaintex
let g:tex_flavor='latex'

" vim-latex compiles to pdf
let g:Tex_DefaultTargetFormat='pdf'

" Redefine g:TexLeader to '#' (from '`') because it's annoying.
let g:TexLeader='#'

" syntax highlighting
syntax on
" color scheme
"colorscheme molokai

" default linewrap
set tw=229

" spellcheck
"set spell

" line numbers
set number
" column numbers
set ruler


" name window after vim file being edited
autocmd BufReadPost,FileReadPost,BufNewFile,BufEnter * call system("tmux rename-window 'vim:" . expand("%:t") ."'")
autocmd VimLeave * call system("tmux setw automatic-rename")

" Make backspace work like normal
set backspace=indent,eol,start

" Plugins moved to nvim config! Raw vim will not use them.

" pane navigation ctrl+hjkl
nnoremap <silent> <C-H> <C-W><C-H>
nnoremap <silent> <C-J> <C-W><C-J>
nnoremap <silent> <C-K> <C-W><C-K>
nnoremap <silent> <C-L> <C-W><C-L>

" buffer navigation alt+hjkl
nnoremap <silent> <A-h> :bprev<CR>
nnoremap <silent> <A-l> :bnext<CR>

" create  newly created windows on the right
set splitright
set splitbelow

