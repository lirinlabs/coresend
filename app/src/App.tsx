import { BrowserRouter, Routes, Route } from 'react-router-dom';
import Landing from './pages/Landing';
import Gateway from './pages/Gateway';
import Inbox from './pages/Inbox';
import NotFound from './pages/NotFound';
import { ThemeProvider } from './components/theme-provider';
import Test from './pages/Test';

const App = () => {
    return (
        <ThemeProvider defaultTheme='system' storageKey='coresend-theme'>
            <BrowserRouter>
                <Routes>
                    <Route path='/' element={<Landing />} />
                    <Route path='/gateway' element={<Gateway />} />
                    <Route path='/inbox' element={<Inbox />} />

                    <Route path='/test' element={<Test />} />
                    <Route path='*' element={<NotFound />} />
                </Routes>
            </BrowserRouter>
        </ThemeProvider>
    );
};

export default App;
